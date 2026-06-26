// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru

import (
	"time"

	"github.com/justpranavrs/tlru/lruclock"
	"github.com/justpranavrs/tlru/lrucore"
	"github.com/justpranavrs/tlru/mux"
)

// PoolTLRU is the better implementation of [lrucore.TLRU]. It creates
// many instances of [lrucore.TLRU] and works based on [LRU].
// It manages a unified clock for all the separate instances.
type PoolTLRU[K comparable, V any] struct {
	pool[K, V, *lrucore.TLRU[K, V]]
	clock *lruclock.Clock
}

// tlruConfig represents the configuration of [TLRU]. It should be used with [TLRUOption].
type tlruConfig struct {
	lruConfig
	clock   *lruclock.Clock
	sliding bool
}

// TLRUOption is used to configure [TLRU] when creating an instance using [NewWithTTL] constructor.
type TLRUOption interface {
	apply(c *tlruConfig) error
}

// tlruOpt represents [TLRU] only options.
type tlruOpt func(c *tlruConfig) error

// apply is an adapter from [tlruOpt] to [TLRUOption].
func (f tlruOpt) apply(c *tlruConfig) error {
	return f(c)
}

// apply is an adapter from [LRUOption] to [TLRUOption].
func (f LRUOption) apply(c *tlruConfig) error {
	return f(&c.lruConfig)
}

// NewWithTTL creates a [TLRU] instance with the given capacity, ttl and options. It creates
// the required [lrucore.TLRU] instances, initiates the [mux.Mux] for shard routing.
// It defaults to the Mux with hash/maphash algorithm. Check `tlru/mux` package for alternatives.
//
// The ttl value is rounded off in terms of its internal clock ticks. Check [lruclock.Clock.Ticks].
//
// It operates on a default clock with 100ms. To customize the
// Clock, refer [WithClock].
//
// Returns [ErrInvalidShards] if shards is not in range [1, 1073741823].
//
// Returns [ErrInvalidCapacity] if capacity is not in the range of int32
// and greater than or equal to twice the number of shards.
func NewWithTTL[K comparable, V any](capacity int, ttl time.Duration, opts ...TLRUOption) (*PoolTLRU[K, V], error) {
	// build the config
	cfg := tlruConfig{
		lruConfig: lruConfig{
			shards: DefaultShards,
			mux:    nil,
		},
		clock: nil,
	}
	for _, opt := range opts { // options
		if opt == nil {
			continue
		}
		if err := opt.apply(&cfg); err != nil {
			return nil, err
		}
	}

	// set the mux hash
	var hash mux.Mux[K]
	if cfg.mux != nil {
		if fun, ok := cfg.mux.(mux.Mux[K]); ok {
			hash = fun
		} else {
			return nil, ErrInvalidMuxKey
		}
	} else {
		hash = mux.NewMH32[K](cfg.shards)
	}

	if cfg.clock == nil {
		cfg.clock = lruclock.New(100 * time.Millisecond)
		_ = cfg.clock.Start()
	}

	createShard := func(cap int) (*lrucore.TLRU[K, V], error) {
		if cfg.sliding {
			return lrucore.NewWithTTL[K, V](cap, ttl, lrucore.WithClock(cfg.clock), lrucore.WithSliding())
		}
		return lrucore.NewWithTTL[K, V](cap, ttl, lrucore.WithClock(cfg.clock))
	}
	pool, err := assemble(capacity, cfg.shards, hash, createShard)
	if err != nil {
		return nil, err
	}

	return &PoolTLRU[K, V]{
		pool:  pool,
		clock: cfg.clock,
	}, nil
}

// WithClock allows the usage of a custom clock for [PoolTLRU].
// It is only initialized if "TTL" is enabled.
//
// NOTE: Using WithClock on [NewWithTTL] will not start the clock. Use [lruclock.Clock.Start] to
// initiate the timer.
func WithClock(clock *lruclock.Clock) tlruOpt {
	return func(c *tlruConfig) error {
		c.clock = clock
		return nil
	}
}

// WithSliding enables Sliding TTL on the LRU cache.
//
// It will update the timestamp of the key on [PoolTLRU.Get] and
// [PoolTLRU.Put].
func WithSliding() tlruOpt {
	return func(c *tlruConfig) error {
		c.sliding = true
		return nil
	}
}

// Close safely closes the background clock when TTL is enabled on the cache.
func (l *PoolTLRU[K, V]) Close() {
	if l.clock != nil {
		l.clock.Stop()
	}
}

// Check [PoolTLRU.Get] on how Get works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *PoolTLRU[K, V]) GetWithTTL(key K) (V, time.Duration, bool) {
	shard := l.mux(key)
	return l.pool.shards[shard].GetWithTTL(key)
}

// Check [PoolTLRU.Peek] on how Peek works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *PoolTLRU[K, V]) PeekWithTTL(key K) (V, time.Duration, bool) {
	shard := l.mux(key)
	return l.pool.shards[shard].PeekWithTTL(key)
}

// PutWithTTL adds a new value to the cache with the given key and the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [PoolTLRU.Put] on how Put works.
func (l *PoolTLRU[K, V]) PutWithTTL(key K, value V, ttl time.Duration) {
	shard := l.mux(key)
	l.pool.shards[shard].PutWithTTL(key, value, ttl)
}

// Refresh resets the TTL of an existing key using the default ttl.
// It returns false if the key could not be found.
func (l *PoolTLRU[K, V]) Refresh(key K) bool {
	shard := l.mux(key)
	return l.pool.shards[shard].Refresh(key)
}

// SetTTL resets the TTL of an existing key using the provided ttl value.
// It returns false if the key could not be found.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (l *PoolTLRU[K, V]) SetTTL(key K, ttl time.Duration) bool {
	shard := l.mux(key)
	return l.pool.shards[shard].SetTTL(key, ttl)
}

// TTL returns the remaining TTL for the key.
func (l *PoolTLRU[K, V]) TTL(key K) (time.Duration, bool) {
	shard := l.mux(key)
	return l.pool.shards[shard].TTL(key)
}

// UpsertWithTTL adds a new value to the cache with the given key and the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [PoolTLRU.Upsert] on how Upsert works.
func (l *PoolTLRU[K, V]) UpsertWithTTL(key K, value V, ttl time.Duration) (lrucore.UpsertState, V) {
	shard := l.mux(key)
	return l.pool.shards[shard].UpsertWithTTL(key, value, ttl)
}

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru

import (
	"time"

	"github.com/justpranavrs/tlru/clock"
	"github.com/justpranavrs/tlru/core"
	"github.com/justpranavrs/tlru/mux"
)

// PoolTLRU is the TTL implementation of [core.TLRU]. It creates
// many instances of [core.TLRU] and works based on [PoolLRU].
// It manages a unified clock for all the separate instances.
type PoolTLRU[K comparable, V any] struct {
	tPool[K, V, *core.TLRU[K, V]]
	clock *clock.Clock
}

// tlruConfig represents the configuration of [PoolTLRU]. It should be used with [TLRUOption].
type tlruConfig struct {
	lruConfig
	clock   *clock.Clock
	sliding bool
}

// TLRUOption is used to configure [PoolTLRU] when creating an instance using [NewWithTTL] constructor.
type TLRUOption interface {
	apply(c *tlruConfig) error
}

// tlruOpt represents [PoolTLRU] only options.
type tlruOpt func(c *tlruConfig) error

// apply is an adapter from [tlruOpt] to [TLRUOption].
func (f tlruOpt) apply(c *tlruConfig) error {
	return f(c)
}

// apply is an adapter from [LRUOption] to [TLRUOption].
func (f LRUOption) apply(c *tlruConfig) error {
	return f(&c.lruConfig)
}

// NewWithTTL creates a [PoolTLRU] instance with the given capacity, ttl and options. It creates
// the required [core.TLRU] instances, initiates the [mux.Mux] for shard routing.
// It defaults to the Mux with hash/maphash algorithm. Check `tlru/mux` package for alternatives.
//
// The ttl value is rounded off in terms of its internal clock ticks. Check [clock.Clock.Ticks].
//
// It operates on a default clock with 100ms. To customize the
// Clock, refer [WithClock].
//
// It has compatibility with [LRUOption] too.
//
// Returns [ErrInvalidShards] if shards is not in range [1, 1000000000].
//
// Returns [ErrInvalidCapacity] if capacity is too small to be configured for the cache.
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
		cfg.clock = clock.New(100 * time.Millisecond)
		_ = cfg.clock.Start()
	}

	createShard := func(cap int) (*core.TLRU[K, V], error) {
		if cfg.sliding {
			return core.NewWithTTL[K, V](cap, ttl, core.WithClock(cfg.clock), core.WithSliding())
		}
		return core.NewWithTTL[K, V](cap, ttl, core.WithClock(cfg.clock))
	}
	pool, err := assembleWithTTL(capacity, cfg.shards, hash, cfg.clock, cfg.sliding, createShard)
	if err != nil {
		return nil, err
	}

	return &PoolTLRU[K, V]{
		tPool: pool,
		clock: cfg.clock,
	}, nil
}

// WithClock allows the usage of a custom clock for [PoolTLRU].
// It is only initialized if "TTL" is enabled.
//
// NOTE: Using WithClock on [NewWithTTL] will not start the clock. Use [clock.Clock.Start] to
// initiate the timer.
func WithClock(clock *clock.Clock) TLRUOption {
	clockOpt := func(c *tlruConfig) error {
		c.clock = clock
		return nil
	}
	return tlruOpt(clockOpt)
}

// WithSliding enables Sliding TTL on the LRU cache.
//
// It will update the timestamp of the key on [PoolTLRU.Get] and
// [PoolTLRU.Put].
func WithSliding() TLRUOption {
	slideOpt := func(c *tlruConfig) error {
		c.sliding = true
		return nil
	}
	return tlruOpt(slideOpt)
}

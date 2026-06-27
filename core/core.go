// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"errors"
	"time"

	"github.com/justpranavrs/tlru/clock"
)

// LRU is basic implementation of 'Least Recently Used' Cache.
// It uses a contiguous array of nodes as opposed to the standard doubly-linked list.
// It uses a hash map to track the key to the index in the array.
//
// It is recommended for key K to be primitives such as ([int], [uint64], [string]).
type LRU[K comparable, V any] struct {
	syncBase[K, V, *base[K, V]]
}

// TLRU is the implementation of 'LRU' with TTL (Time-To-Live). It
// operates on an internal clock from [clock.Clock] and operates with an instance
// of [LRU].
type TLRU[K comparable, V any] struct {
	syncBase[K, V, *tlruBase[K, V]]
}

// ttlConfig represents the configuration of [TLRU]. It should be used with [TTLOption].
type ttlConfig struct {
	// internal clock
	clock *clock.Clock

	// sliding TTL
	sliding bool
}

// TTLOption is used to configure [TLRU] when creating an instance using [NewWithTTL] constructor.
type TTLOption func(c *ttlConfig)

var (
	// ErrInvalidBatchSize is returned by batch operations when keys and values do not have the same lengths.
	ErrInvalidBatchSize = errors.New("invalid LRU batch sizes: keys and values do not have the same lengths")

	// ErrInvalidCapacity is returned by [New] or [NewWithTTL] when the maximum cache capacity is not in [2, 2147483645].
	ErrInvalidCapacity = errors.New("invalid LRU cache capacity: must be in the range [2, 2147483645]")
)

// New creates an instance of [LRU] using the given capacity.
//
// Returns an [ErrInvalidCapacity] if the capacity is not in [2, 2147483645].
func New[K comparable, V any](capacity int) (*LRU[K, V], error) {
	lru, err := assembleBase[K, V](capacity)
	if err != nil {
		return nil, err
	}
	return &LRU[K, V]{
		syncBase: syncBase[K, V, *base[K, V]]{
			lru: lru,
		},
	}, nil
}

// NewWithTTL creates an instance of [TLRU] using the given capacity and sets
// the default expiration timer based on the argument "ttl".
//
// The ttl value is rounded off in terms of its internal clock ticks.
// Check [clock.Clock.Ticks].
//
// It operates on a default clock with 100ms. To customize the
// Clock, refer [WithClock].
//
// Returns an [ErrInvalidCapacity] if the capacity is not in [2, 2147483645].
func NewWithTTL[K comparable, V any](capacity int, ttl time.Duration, opts ...TTLOption) (*TLRU[K, V], error) {
	// build the config
	cfg := ttlConfig{
		clock:   nil,
		sliding: false,
	}
	for _, opt := range opts { // options
		if opt == nil {
			continue
		}
		opt(&cfg)
	}

	var clk *clock.Clock
	if cfg.clock != nil {
		clk = cfg.clock
	} else {
		clk = clock.New(100 * time.Millisecond)
		_ = clk.Start()
	}

	lru, err := assembleTLRU[K, V](capacity, ttl, clk, cfg.sliding)
	if err != nil {
		return nil, err
	}
	return &TLRU[K, V]{
		syncBase: syncBase[K, V, *tlruBase[K, V]]{
			lru: lru,
		},
	}, nil
}

// WithClock allows the usage of a custom clock for [TLRU].
//
// NOTE: Using WithClock on [NewWithTTL] will not start the clock. Use [clock.Clock.Start] to
// initiate the timer.
func WithClock(clock *clock.Clock) TTLOption {
	return func(c *ttlConfig) {
		c.clock = clock
	}
}

// WithSliding enables Sliding TTL on the LRU cache.
//
// It will update the timestamp of the key on [TLRU.Get] and
// [TLRU.Put] using the TTL provided in [NewWithTTL].
func WithSliding() TTLOption {
	return func(c *ttlConfig) {
		c.sliding = true
	}
}

// Close safely closes the background clock when TTL is enabled on the cache.
func (t *TLRU[K, V]) Close() {
	t.lru.Close()
}

// Check [TLRU.Get] on how Get works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (t *TLRU[K, V]) GetWithTTL(key K) (V, time.Duration, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.GetWithTTL(key)
}

// Check [TLRU.Peek] on how Peek works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (t *TLRU[K, V]) PeekWithTTL(key K) (V, time.Duration, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.PeekWithTTL(key)
}

// PutWithTTL adds a new value to the cache with the given key and the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [TLRU.Put] on how Put works.
func (t *TLRU[K, V]) PutWithTTL(key K, value V, ttl time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lru.PutWithTTL(key, value, ttl)
}

// Refresh resets the TTL of an existing key using the default ttl.
// It returns false if the key could not be found.
func (t *TLRU[K, V]) Refresh(key K) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.Refresh(key)
}

// SetTTL resets the TTL of an existing key using the provided ttl value.
// It returns false if the key could not be found.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (t *TLRU[K, V]) SetTTL(key K, ttl time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.SetTTL(key, ttl)
}

// TTL returns the remaining TTL for the key.
func (t *TLRU[K, V]) TTL(key K) (time.Duration, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.TTL(key)
}

// UpsertWithTTL adds a new value to the cache with the given key and the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [TLRU.Upsert] on how Upsert works.
func (t *TLRU[K, V]) UpsertWithTTL(key K, value V, ttl time.Duration) (UpsertState, V) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.UpsertWithTTL(key, value, ttl)
}

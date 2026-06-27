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
	syncBase[K, V, *lruBase[K, V]]
}

// SLRU is the implementation of Segmented LRU. It uses two LRU instances,
// probationary and protected to reduce sequential scan problems from
// a single LRU instance.
type SLRU[K comparable, V any] struct {
	syncBase[K, V, *slruBase[K, V]]
}

// TLRU is the implementation of 'LRU' with TTL (Time-To-Live). It
// operates on an internal clock from [clock.Clock] and operates with an instance
// of [LRU].
type TLRU[K comparable, V any] struct {
	tSyncBase[K, V, *tlruBase[K, V, *lruBase[K, ttlValue[V]]]]
}

// TSLRU is the implementation of 'SLRU' with TTL (Time-To-Live). It
// operates on an internal clock from [clock.Clock] and operates with an instance
// of [SLRU].
type TSLRU[K comparable, V any] struct {
	tSyncBase[K, V, *tlruBase[K, V, *slruBase[K, ttlValue[V]]]]
}

// ttlConfig represents the configuration of [TLRU] and [TSLRU]. It should be used with [TTLOption].
type ttlConfig struct {
	// internal clock
	clock *clock.Clock

	// sliding TTL
	sliding bool
}

// TTLOption is used to configure [TLRU] or [TSLRU] when creating an instance using their constructors.
type TTLOption func(c *ttlConfig)

var (
	// ErrInvalidBatchSize is returned by batch operations when keys and values do not have the same lengths.
	ErrInvalidBatchSize = errors.New("invalid batch sizes: keys and values do not have the same lengths")

	// ErrInvalidCapacity is returned by constructors when the maximum cache capacity is too small to be configured for the cache.
	ErrInvalidCapacity = errors.New("invalid cache capacity: total capacity is too small for the configured for the cache")

	// ErrInvalidSLRURatio is returned by [NewSLRU] when the ratio is not between 0 and 100.
	ErrInvalidSLRURatio = errors.New("invalid probationary ratio: value must be between 0 and 100")
)

// New creates an instance of [LRU] using the given capacity.
//
// Returns an [ErrInvalidCapacity] if the capacity is too small to be configured for the cache.
func New[K comparable, V any](capacity int) (*LRU[K, V], error) {
	lru, err := assembleLRU[K, V](capacity)
	if err != nil {
		return nil, err
	}
	return &LRU[K, V]{
		syncBase: syncBase[K, V, *lruBase[K, V]]{
			lru: lru,
		},
	}, nil
}

// NewSLRU creates an instance of [SLRU] using the given capacity.
// It takes in a ratio(0-100) which declares the ratio of probationary capacity
// to the capacity. It is in Percentage.
//
// Example : If the ratio is 5 and the capacity is 1000,
// Capacity :- Probationary : 50, Protected : 950
//
// Returns an [ErrInvalidCapacity] if each of the probationary
// and protected capacities are too small to be configured for the caches.
//
// Returns an [ErrInvalidSLRURatio] if the ratio is not between 0 and 100.
func NewSLRU[K comparable, V any](capacity int, ratio int) (*SLRU[K, V], error) {
	lru, err := assembleSLRU[K, V](capacity, ratio)
	if err != nil {
		return nil, err
	}
	return &SLRU[K, V]{
		syncBase: syncBase[K, V, *slruBase[K, V]]{
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
// Returns an [ErrInvalidCapacity] if the capacity is too small to be configured for the cache.
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

	lru, err := assembleLRU[K, ttlValue[V]](capacity)
	if err != nil {
		return nil, err
	}
	return &TLRU[K, V]{
		tSyncBase: tSyncBase[K, V, *tlruBase[K, V, *lruBase[K, ttlValue[V]]]]{
			lru: assembleWithTTL(lru, ttl, clk, cfg.sliding),
		},
	}, nil
}

// NewSLRUWithTTL creates an instance of [TSLRU] using the given capacity and sets
// the default expiration timer based on the argument "ttl".
//
// Refer [NewSLRU] for more details.
//
// Refer [NewWithTTL] for more details.
func NewSLRUWithTTL[K comparable, V any](capacity int, ratio int, ttl time.Duration, opts ...TTLOption) (*TSLRU[K, V], error) {
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

	lru, err := assembleSLRU[K, ttlValue[V]](capacity, ratio)
	if err != nil {
		return nil, err
	}
	return &TSLRU[K, V]{
		tSyncBase: tSyncBase[K, V, *tlruBase[K, V, *slruBase[K, ttlValue[V]]]]{
			lru: assembleWithTTL(lru, ttl, clk, cfg.sliding),
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

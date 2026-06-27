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

// PoolSLRU is the better implementation of [core.SLRU]. It is a
// 'Least Recently Used' cache with many instances of [core.SLRU]
// to prevent mutual extension locks on a single instance.
//
// Refer [PoolLRU] for more implementation details.
type PoolSLRU[K comparable, V any] struct {
	pool[K, V, *core.SLRU[K, V]]
}

// PoolTSLRU is the TTL implementation of [core.TSLRU]. It creates
// many instances of [core.TSLRU] and works based on [PoolSLRU].
// It manages a unified clock for all the separate instances.
type PoolTSLRU[K comparable, V any] struct {
	tPool[K, V, *core.TSLRU[K, V]]
	clock *clock.Clock
}

// NewSLRU creates a [PoolSLRU] instance with the given capacity and options. It creates
// the required [core.SLRU] instances, initiates the [mux.Mux] for shard routing.
// It defaults to the Mux with hash/maphash algorithm. Check `tlru/mux` package for alternatives.
// It takes in a ratio(0-100) which declares the ratio of probationary capacity
// to the capacity. It is in Percentage.
//
// Example : If the ratio is 5 and the capacity is 1000,
// Capacity :- Probationary : 50, Protected : 950
//
// It has compatibility with [LRUOption] too.
//
// Returns [ErrInvalidShards] if shards is not in range [1, 1000000000].
//
// Returns [ErrInvalidCapacity] if capacity is too small to be configured for the cache.
//
// Returns an [ErrInvalidSLRURatio] if the ratio is not between 0 and 100.
func NewSLRU[K comparable, V any](capacity int, ratio int, opts ...LRUOption) (*PoolSLRU[K, V], error) {
	// build the config
	cfg := lruConfig{
		shards: DefaultShards,
		mux:    nil,
	}
	for _, opt := range opts { // options
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
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

	createShard := func(cap int) (*core.SLRU[K, V], error) {
		return core.NewSLRU[K, V](cap, ratio)
	}
	pool, err := assemble(capacity, cfg.shards, hash, createShard)
	if err != nil {
		return nil, err
	}
	return &PoolSLRU[K, V]{
		pool: pool,
	}, nil
}

// NewSLRUWithTTL creates a [PoolTSLRU] instance with the given capacity and options. It creates
// the required [core.TSLRU] instances, initiates the [mux.Mux] for shard routing.
//
// Refer [NewSLRU] for more details.
//
// Refer [NewWithTTL] for more details.
func NewSLRUWithTTL[K comparable, V any](capacity int, ratio int, ttl time.Duration, opts ...TLRUOption) (*PoolTSLRU[K, V], error) {
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

	createShard := func(cap int) (*core.TSLRU[K, V], error) {
		if cfg.sliding {
			return core.NewSLRUWithTTL[K, V](cap, ratio, ttl, core.WithClock(cfg.clock), core.WithSliding())
		}
		return core.NewSLRUWithTTL[K, V](cap, ratio, ttl, core.WithClock(cfg.clock))
	}
	pool, err := assembleWithTTL(capacity, cfg.shards, hash, cfg.clock, cfg.sliding, createShard)
	if err != nil {
		return nil, err
	}

	return &PoolTSLRU[K, V]{
		tPool: pool,
		clock: cfg.clock,
	}, nil
}

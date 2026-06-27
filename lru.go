// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru

import (
	"math"

	core "github.com/justpranavrs/tlru/core"
	"github.com/justpranavrs/tlru/mux"
)

// DefaultShards represents the number of shards allocated to LRU
// if [WithShards] option is not configured.
const DefaultShards int = 128

// PoolLRU is the better implementation of [core.LRU]. It is a
// 'Least Recently Used' cache with many instances of [core.LRU]
// to prevent mutual extension locks on a single instance.
// It has thread-safe but not a completely lock free operations.
//
// It doesn't work on the standard principle of LRU, rather when the mux
// routes it to a [core.LRU] instance. LRU works on shard-local based eviction
// not on globally oldest item in the cache.
//
// [mux.Mux] takes care of routing the shards to their containers consistently
// using its hashing algorithm. The default mux is [mux.NewMH32].
type PoolLRU[K comparable, V any] struct {
	pool[K, V, *core.LRU[K, V]]
}

// lruConfig represents the configuration of [PoolLRU]. It should be used with [LRUOption].
type lruConfig struct {
	shards int
	mux    any
}

// LRUOption is used to configure [PoolLRU] when creating an instance using [New] constructor.
type LRUOption func(c *lruConfig) error

// New creates a [PoolLRU] instance with the given capacity and options. It creates
// the required [core.PoolLRU] instances, initiates the [mux.Mux] for shard routing.
// It defaults to the Mux with hash/maphash algorithm. Check `tlru/mux` package for alternatives.
//
// Returns [ErrInvalidShards] if shards is not in range [1, 1073741823].
//
// Returns [ErrInvalidCapacity] if capacity is not in the range of int32
// and greater than or equal to twice the number of shards.
func New[K comparable, V any](capacity int, opts ...LRUOption) (*PoolLRU[K, V], error) {
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

	createShard := func(cap int) (*core.LRU[K, V], error) {
		return core.New[K, V](cap)
	}
	pool, err := assemble(capacity, cfg.shards, hash, createShard)
	if err != nil {
		return nil, err
	}
	return &PoolLRU[K, V]{
		pool: pool,
	}, nil
}

// WithMux requires a custom [mux.Mux] type function. It is used
// with [PoolLRU] to configure its mux, which is responsible for routing the shards.
func WithMux[K comparable](cm mux.Mux[K]) LRUOption {
	return func(c *lruConfig) error {
		c.mux = cm
		return nil
	}
}

// WithShards assigns the [PoolLRU] instance with num shards. Shards are separate instances
// of [core.Shard] to prevent mutex locks from slowing down the cache.
//
// Returns [ErrInvalidShards] if num is not in range [1, 1073741823].
func WithShards(num int) LRUOption {
	return func(c *lruConfig) error {
		if num < 1 || (num > (math.MaxInt32 >> 1)) {
			return ErrInvalidShards
		}
		c.shards = num
		return nil
	}
}

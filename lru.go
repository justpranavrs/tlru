// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru

import (
	"math"

	"github.com/justpranavrs/tlru/lrucore"
	"github.com/justpranavrs/tlru/mux"
)

// DefaultShards represents the number of shards allocated to LRU
// if [WithShards] option is not configured.
const DefaultShards int = 128

type LRU[K comparable, V any] struct {
	cache coreCluster[K, V, *lrucore.Core[K, V]]
}

// lruConfig represents the configuration of [LRU]. It should be used with [LRUOption].
type lruConfig struct {
	shards int
	mux    any
}

// LRUOption is used to configure [LRU] when creating an instance using [New] constructor.
type LRUOption func(c *lruConfig) error

// New creates a [LRU] instance with the given capacity and options. It creates
// the required [lrucore.Core] instances, initiates the [mux.Mux] for shard routing.
// It defaults to the Mux with hash/maphash algorithm. Check `tlru/mux` package for alternatives.
//
// Returns [ErrInvalidShards] if shards is not in range [1, 1073741823].
//
// Returns [ErrInvalidCapacity] if capacity is not in the range of int32
// and greater than or equal to twice the number of shards.
func New[K comparable, V any](capacity int, opts ...LRUOption) (*LRU[K, V], error) {
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

	cache, err := buildCluster(capacity, cfg.shards, hash, func(cap int) (*lrucore.Core[K, V], error) {
		return lrucore.New[K, V](cap)
	})
	if err != nil {
		return nil, err
	}
	return &LRU[K, V]{
		cache: cache,
	}, nil
}

// WithMux requires a custom [mux.Mux] type function. It is used
// with [LRU] to configure its mux, which is responsible for routing the shards.
func WithMux[K comparable](cm mux.Mux[K]) LRUOption {
	return func(c *lruConfig) error {
		c.mux = cm
		return nil
	}
}

// WithShards assigns the [LRU] instance with num shards. Shards are separate instances
// of [lrucore.Shard] to prevent mutex locks from slowing down the cache.
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

// Capacity returns the maximum allocated capacity of the LRU cache
// across all sharded instances of [lrucore.TTLCore].
func (l *LRU[K, V]) Capacity() int {
	return l.cache.capacity
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *LRU[K, V]) Delete(key K) (V, bool) {
	return l.cache.Delete(key)
}

// Flush clears the LRU cache of all its keys and values across
// all sharded instances.
func (l *LRU[K, V]) Flush() {
	l.cache.Flush()
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
// It updates the key as 'recent' only in its respective shard.
func (l *LRU[K, V]) Get(key K) (V, bool) {
	return l.cache.Get(key)
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found.
func (l *LRU[K, V]) Peek(key K) (V, bool) {
	return l.cache.Peek(key)
}

// Put adds a new value to the cache with the given key.
// It updates the key as 'recent' only in its respective shard.
// It evicts the key only from the respective shard the key is linked to.
func (l *LRU[K, V]) Put(key K, value V) {
	l.Upsert(key, value)
}

// ResetStats resets the stats of the sharded LRU cache.
func (l *LRU[K, V]) ResetStats() {
	l.cache.ResetStats()
}

// Shards returns the number of sharded instances in the LRU cache.
func (l *LRU[K, V]) Shards() int {
	return l.cache.Shards()
}

// Size returns the current size of the LRU cache
// across all sharded instances.
func (l *LRU[K, V]) Size() int {
	return l.cache.Size()
}

// Stats return the current stats of the sharded LRU cache.
func (l *LRU[K, V]) Stats() lrucore.CoreStats {
	return l.cache.Stats()
}

// Upsert adds a new value to the cache with the given key.
// It returns a value based on how the internal state of the cache changed.
// It evicts or updates locally on the shard, instead of global cache.
// Returns a value [lrucore.UpsertState].
func (l *LRU[K, V]) Upsert(key K, value V) (lrucore.UpsertState, V) {
	return l.cache.Upsert(key, value)
}

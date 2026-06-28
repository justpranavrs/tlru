// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru

import (
	"errors"
	"math"

	"github.com/justpranavrs/tlru/core"
	"github.com/justpranavrs/tlru/mux"
)

// Pool defines the general implementation of a 'Least Recently Used' cache
// with a collection of [core.Shard].
// It has thread-safe operations.
//
// K represents the type of the key, whereas V represents the type of the Value
// in the cache
type Pool[K comparable, V any] interface {
	// Capacity returns the maximum allocated capacity of the LRU cache.
	Capacity() int

	// Close safely terminates the Pool instance.
	Close()

	// Contains checks whether the key is present in the LRU cache.
	Contains(key K) bool

	// Delete removes the key from the cache and returns the evicted value.
	// It returns false if the key was not found in the cache.
	Delete(key K) (V, bool)

	// Flush clears the LRU cache of all its keys and values.
	Flush()

	// Get retrieves the cache value using key.
	// It returns false if the key is not found.
	Get(key K) (V, bool)

	// Peek retrieves the cache value without updating it
	// to be the most recently used.
	// It returns false if the key is not found.
	Peek(key K) (V, bool)

	// Put adds a new value to the cache with the given key.
	Put(key K, value V)

	// ResetStats resets the stats of the LRU cache.
	ResetStats()

	// Shards returns the number of sharded instances in the LRU cache.
	Shards() int

	// Size returns the current size of the LRU cache.
	Size() int

	// Stats return the current stats of the LRU cache.
	Stats() core.Stats

	// Upsert adds a new value to the cache with the given key.
	// It returns a value based on how the internal state of the cache changed.
	Upsert(key K, value V) (core.UpsertState, V)
}

// pool is the internal sharding container used by [PoolLRU] and [PoolSLRU].
// It handles routing via [mux.Mux] and maintains the shards.
type pool[K comparable, V any, C core.Shard[K, V]] struct {
	// capacity represents the maximum allocated space for the LRU cache.
	capacity int

	// shards is an array of the multiple instances of [core.Shard].
	shards []C

	// mux is the router for the shards in the shards array.
	// It takes in a key K and outputs a hash [uint32]
	mux mux.Mux[K]
}

var (
	// ErrInvalidCapacity is returned by constructors when the maximum cache capacity is too small for the configured for the cache.
	ErrInvalidCapacity = errors.New("invalid cache capacity: total capacity is too small for the configured for the cache")

	// ErrInvalidMuxKey is returned by constructors when mux does not have the same key type as the cache.
	ErrInvalidMuxKey = errors.New("invalid mux for cache: mux does not have the same key type as the cache")

	// ErrInvalidShards is returned by constructors when an invalid number of shards is passed using WithShards.
	ErrInvalidShards = errors.New("invalid number of shards: must be in range [1, 1000000000]")
)

// assemble creates a [pool] instance with the provided arguments. It creates shards based on the createShard function.
func assemble[K comparable, V any, C core.Shard[K, V]](capacity int, nShards int, hash mux.Mux[K], createShard func(cap int) (C, error)) (pool[K, V, C], error) {
	var zero pool[K, V, C]
	if capacity < int(nShards*2) || (capacity > math.MaxInt32) {
		return zero, ErrInvalidCapacity
	}

	var err error
	shards := make([]C, nShards)

	cap := capacity / nShards // create base instances
	rem := capacity % nShards
	for i := range nShards {
		sCap := cap
		if i < rem {
			sCap++
		}

		shards[i], err = createShard(sCap)
		if err != nil {
			return zero, err
		}
	}
	return pool[K, V, C]{
		capacity: capacity,
		shards:   shards,
		mux:      hash,
	}, nil
}

// Capacity returns the maximum allocated capacity of the LRU cache
// across all sharded instances of [core.Shard].
func (l *pool[K, V, C]) Capacity() int {
	return l.capacity
}

// Close safely terminates the Pool instance and frees up the memory.
func (l *pool[K, V, C]) Close() {
	for _, c := range l.shards {
		c.Close()
	}
}

// Contains checks whether the key is present in the LRU cache.
func (l *pool[K, V, C]) Contains(key K) bool {
	shard := l.mux(key)
	return l.shards[shard].Contains(key)
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *pool[K, V, C]) Delete(key K) (V, bool) {
	shard := l.mux(key)
	value, ok := l.shards[shard].Delete(key)
	if !ok {
		return *new(V), false
	}
	return value, true
}

// Flush clears the LRU cache of all its keys and values across
// all sharded instances.
func (l *pool[K, V, C]) Flush() {
	for _, c := range l.shards {
		c.Flush()
	}
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
// It updates the key as 'recent' only in its respective shard.
func (l *pool[K, V, C]) Get(key K) (V, bool) {
	shard := l.mux(key)
	value, ok := l.shards[shard].Get(key)
	if !ok {
		return *new(V), false
	}
	return value, true
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found.
func (l *pool[K, V, C]) Peek(key K) (V, bool) {
	shard := l.mux(key)
	value, ok := l.shards[shard].Peek(key)
	if !ok {
		return *new(V), false
	}
	return value, true
}

// Put adds a new value to the cache with the given key.
// It updates the key as 'recent' only in its respective shard.
// It evicts the key only from the respective shard the key is linked to.
func (l *pool[K, V, C]) Put(key K, value V) {
	shard := l.mux(key)
	l.shards[shard].Put(key, value)
}

// ResetStats resets the stats of the sharded LRU cache.
func (l *pool[K, V, C]) ResetStats() {
	for i := range l.shards {
		l.shards[i].ResetStats()
	}
}

// Shards returns the number of sharded instances in the LRU cache.
func (l *pool[K, V, C]) Shards() int {
	return len(l.shards)
}

// Size returns the current size of the LRU cache
// across all sharded instances.
func (l *pool[K, V, C]) Size() int {
	size := 0
	for i := range l.shards {
		size += l.shards[i].Size()
	}
	return size
}

// Stats return the current stats of the sharded LRU cache.
func (l *pool[K, V, C]) Stats() core.Stats {
	stats := core.Stats{}
	for i := range l.shards {
		st := l.shards[i].Stats()
		stats.Hits += st.Hits
		stats.Misses += st.Misses
		stats.Evictions += st.Evictions
		stats.Expirations += st.Expirations
	}
	return stats
}

// Upsert adds a new value to the cache with the given key.
// It returns a value based on how the internal state of the cache changed.
// It evicts or updates locally on the shard, instead of global cache.
// Returns a value [core.UpsertState].
//
// It also returns a value based on [core.UpsertState]
//   - [core.UpsertAddNoEviction] returns the zero value of V.
//   - [core.UpsertAddWithEviction] returns the evicted value.
//   - [core.UpsertReplace] returns the old value the key had.
func (l *pool[K, V, C]) Upsert(key K, value V) (core.UpsertState, V) {
	shard := l.mux(key)
	return l.shards[shard].Upsert(key, value)
}

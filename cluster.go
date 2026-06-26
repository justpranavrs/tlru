// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru

import (
	"errors"
	"math"

	"github.com/justpranavrs/tlru/lrucore"
	"github.com/justpranavrs/tlru/mux"
)

// Cluster defines the general implementation of a 'Least Recently Used' cache
// with a collection of [lrucore.Shard].
// It has thread-safe operations.
//
// K represents the type of the key, whereas V represents the type of the Value
// in the cache
type Cluster[K comparable, V any] interface {
	// Capacity returns the maximum allocated capacity of the LRU cache.
	Capacity() int

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
	Stats() lrucore.CoreStats

	// Upsert adds a new value to the cache with the given key.
	// It returns a value based on how the internal state of the cache changed.
	Upsert(key K, value V) (lrucore.UpsertState, V)
}

// coreCluster is the better implementation of [lrucore.Shard]. It is a
// 'Least Recently Used' cache with many instances of [lrucore.Shard]
// to prevent mutual extension locks on a single instance.
// It has thread-safe but not a completely lock free operations.
//
// It doesn't work on the standard principle of LRU, rather when the mux
// routes it to a [lrucore.Shard] instance. LRU works on shard-local based eviction
// not on globally oldest item in the cache.
//
// [mux.Mux] takes care of routing the shards to their containers consistently
// using its hashing algorithm. The default mux is [mux.NewMH32].
type coreCluster[K comparable, V any, C lrucore.Shard[K, V]] struct {
	// capacity represents the maximum allocated space for the LRU cache.
	capacity int

	// shards is an array of the multiple instances of [lrucore.Shard].
	shards []C

	// mux is the router for the shards in the shards array.
	// It takes in a key K and outputs a hash [uint32]
	mux mux.Mux[K]
}

var (
	// ErrInvalidCapacity is returned by [New] when an invalid cache capacity is passed as argument.
	ErrInvalidCapacity = errors.New("invalid LRU cache capacity: must be in the range of int32 and greater than or equal to twice the number of shards")

	// ErrInvalidMuxKey is returned by [New] when mux does not have the same key type as LRU.
	ErrInvalidMuxKey = errors.New("invalid mux for LRU: mux does not have the same key type as LRU")

	// ErrInvalidShards is returned by [New] when an invalid number of shards is passed using WithShards.
	ErrInvalidShards = errors.New("invalid number of shards: must be in range [1, 1073741823]")
)

// buildCluster creates a [cache] instance with the provided arguments. It creates shards based on the createShard function.
func buildCluster[K comparable, V any, C lrucore.Shard[K, V]](capacity int, nShards int, hash mux.Mux[K], createShard func(cap int) (C, error)) (coreCluster[K, V, C], error) {
	var zero coreCluster[K, V, C]
	if capacity < int(nShards*2) || (capacity > math.MaxInt32) {
		return zero, ErrInvalidCapacity
	}

	var err error
	shards := make([]C, nShards)

	cap := capacity / nShards // create lrucore instances
	rem := capacity % nShards
	for i := range nShards {
		sCap := cap
		if i < rem {
			sCap++
		}

		shards[i], err = createShard(sCap)
		if err != nil {
			if errors.Is(err, lrucore.ErrInvalidCapacity) {
				return zero, ErrInvalidCapacity
			}
			return zero, err
		}
	}
	return coreCluster[K, V, C]{
		capacity: capacity,
		shards:   shards,
		mux:      hash,
	}, nil
}

// Capacity returns the maximum allocated capacity of the LRU cache
// across all sharded instances of [lrucore.Core].
func (l *coreCluster[K, V, C]) Capacity() int {
	return l.capacity
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *coreCluster[K, V, C]) Delete(key K) (V, bool) {
	shard := l.mux(key)
	value, ok := l.shards[shard].Delete(key)
	if !ok {
		return *new(V), false
	}
	return value, true
}

// Flush clears the LRU cache of all its keys and values across
// all sharded instances.
func (l *coreCluster[K, V, C]) Flush() {
	for _, c := range l.shards {
		c.Flush()
	}
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
// It updates the key as 'recent' only in its respective shard.
func (l *coreCluster[K, V, C]) Get(key K) (V, bool) {
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
func (l *coreCluster[K, V, C]) Peek(key K) (V, bool) {
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
func (l *coreCluster[K, V, C]) Put(key K, value V) {
	l.Upsert(key, value)
}

// ResetStats resets the stats of the sharded LRU cache.
func (l *coreCluster[K, V, C]) ResetStats() {
	for i := range l.shards {
		l.shards[i].ResetStats()
	}
}

// Shards returns the number of sharded instances in the LRU cache.
func (l *coreCluster[K, V, C]) Shards() int {
	return len(l.shards)
}

// Size returns the current size of the LRU cache
// across all sharded instances.
func (l *coreCluster[K, V, C]) Size() int {
	size := 0
	for i := range l.shards {
		size += l.shards[i].Size()
	}
	return size
}

// Stats return the current stats of the sharded LRU cache.
func (l *coreCluster[K, V, C]) Stats() lrucore.CoreStats {
	stats := lrucore.CoreStats{}
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
// Returns a value [lrucore.UpsertState].
func (l *coreCluster[K, V, C]) Upsert(key K, value V) (lrucore.UpsertState, V) {
	shard := l.mux(key)
	return l.shards[shard].Upsert(key, value)
}

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru

import (
	"math"
	"time"

	"github.com/justpranavrs/tlru/clock"
	"github.com/justpranavrs/tlru/core"
	"github.com/justpranavrs/tlru/mux"
)

// TPool defines the general implementation of a 'Least Recently Used' cache
// with a collection of [core.TTLShard].
// It has thread-safe operations.
//
// K represents the type of the key, whereas V represents the type of the Value
// in the cache
type TPool[K comparable, V any] interface {
	Pool[K, V]

	// Close safely closes the background clock when TTL is enabled on the cache.
	Close()

	// Check [Pool.Get] on how Get works.
	// It also returns the remaining TTL in the key if it was found in the cache.
	GetWithTTL(key K) (V, time.Duration, bool)

	// Check [Pool.Peek] on how Peek works.
	// It also returns the remaining TTL in the key if it was found in the cache.
	PeekWithTTL(key K) (V, time.Duration, bool)

	// PutWithTTL adds a new value to the cache with the provided ttl value.
	PutWithTTL(key K, value V, ttl time.Duration)

	// Refresh resets the TTL of an existing key using the default ttl.
	// It returns false if the key could not be found.
	Refresh(key K) bool

	// SetTTL resets the TTL of an existing key using the given ttl in the argument.
	// It returns false if the key could not be found.
	SetTTL(key K, ttl time.Duration) bool

	// TTL returns the remaining TTL for the key.
	TTL(key K) (time.Duration, bool)

	// UpsertWithTTL adds a new value to the cache with the given key and the provided ttl value.
	//
	// Check [Pool.Upsert] on how Upsert works.
	UpsertWithTTL(key K, value V, ttl time.Duration) (core.UpsertState, V)
}

// tPool is the internal sharding container used by [PoolTLRU] and [PoolTSLRU].
// It handles routing via [mux.Mux] and maintains the shards.
//
// tPool does not implement pool because GoDoc doesn't recognize multiple
// levels of struct embedding even though compiler can.
type tPool[K comparable, V any, C core.TTLShard[K, V]] struct {
	// capacity represents the maximum allocated space for the LRU cache.
	capacity int

	// shards is an array of the multiple instances of [core.Shard].
	shards []C

	// mux is the router for the shards in the shards array.
	// It takes in a key K and outputs a hash [uint32]
	mux mux.Mux[K]

	// ttl clock and sliding ttl
	clock   *clock.Clock
	sliding bool
}

// assembleWithTTL creates a [tPool] instance with the provided arguments. It creates shards based on the createShard function.
func assembleWithTTL[K comparable, V any, C core.TTLShard[K, V]](capacity int, nShards int,
	hash mux.Mux[K], clock *clock.Clock, sliding bool, createShard func(cap int) (C, error),
) (tPool[K, V, C], error) {
	var zero tPool[K, V, C]
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
	return tPool[K, V, C]{
		capacity: capacity,
		shards:   shards,
		mux:      hash,
		clock:    clock,
		sliding:  sliding,
	}, nil
}

// Close safely closes the background clock when TTL is enabled and also frees up 
// memory on the cache.
func (l *tPool[K, V, C]) Close() {
	if l.clock != nil {
		l.clock.Stop()
	}

	for _, s := range l.shards {
		s.Close()
	}
}

// Check [PoolTLRU.Get] on how Get works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *tPool[K, V, C]) GetWithTTL(key K) (V, time.Duration, bool) {
	shard := l.mux(key)
	return l.shards[shard].GetWithTTL(key)
}

// Check [PoolTLRU.Peek] on how Peek works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *tPool[K, V, C]) PeekWithTTL(key K) (V, time.Duration, bool) {
	shard := l.mux(key)
	return l.shards[shard].PeekWithTTL(key)
}

// PutWithTTL adds a new value to the cache with the given key and the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [PoolTLRU.Put] on how Put works.
func (l *tPool[K, V, C]) PutWithTTL(key K, value V, ttl time.Duration) {
	shard := l.mux(key)
	l.shards[shard].PutWithTTL(key, value, ttl)
}

// Refresh resets the TTL of an existing key using the default ttl.
// It returns false if the key could not be found.
func (l *tPool[K, V, C]) Refresh(key K) bool {
	shard := l.mux(key)
	return l.shards[shard].Refresh(key)
}

// SetTTL resets the TTL of an existing key using the provided ttl value.
// It returns false if the key could not be found.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (l *tPool[K, V, C]) SetTTL(key K, ttl time.Duration) bool {
	shard := l.mux(key)
	return l.shards[shard].SetTTL(key, ttl)
}

// TTL returns the remaining TTL for the key.
func (l *tPool[K, V, C]) TTL(key K) (time.Duration, bool) {
	shard := l.mux(key)
	return l.shards[shard].TTL(key)
}

// UpsertWithTTL adds a new value to the cache with the given key and the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [PoolTLRU.Upsert] on how Upsert works.
func (l *tPool[K, V, C]) UpsertWithTTL(key K, value V, ttl time.Duration) (core.UpsertState, V) {
	shard := l.mux(key)
	return l.shards[shard].UpsertWithTTL(key, value, ttl)
}

// Capacity returns the maximum allocated capacity of the LRU cache
// across all sharded instances of [core.Shard].
func (l *tPool[K, V, C]) Capacity() int {
	return l.capacity
}

// Contains checks whether the key is present in the LRU cache.
func (l *tPool[K, V, C]) Contains(key K) bool {
	shard := l.mux(key)
	return l.shards[shard].Contains(key)
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *tPool[K, V, C]) Delete(key K) (V, bool) {
	shard := l.mux(key)
	value, ok := l.shards[shard].Delete(key)
	if !ok {
		return *new(V), false
	}
	return value, true
}

// Flush clears the LRU cache of all its keys and values across
// all sharded instances.
func (l *tPool[K, V, C]) Flush() {
	for _, c := range l.shards {
		c.Flush()
	}
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
// It updates the key as 'recent' only in its respective shard.
func (l *tPool[K, V, C]) Get(key K) (V, bool) {
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
func (l *tPool[K, V, C]) Peek(key K) (V, bool) {
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
func (l *tPool[K, V, C]) Put(key K, value V) {
	shard := l.mux(key)
	l.shards[shard].Put(key, value)
}

// ResetStats resets the stats of the sharded LRU cache.
func (l *tPool[K, V, C]) ResetStats() {
	for i := range l.shards {
		l.shards[i].ResetStats()
	}
}

// Shards returns the number of sharded instances in the LRU cache.
func (l *tPool[K, V, C]) Shards() int {
	return len(l.shards)
}

// Size returns the current size of the LRU cache
// across all sharded instances.
func (l *tPool[K, V, C]) Size() int {
	size := 0
	for i := range l.shards {
		size += l.shards[i].Size()
	}
	return size
}

// Stats return the current stats of the sharded LRU cache.
func (l *tPool[K, V, C]) Stats() core.Stats {
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
func (l *tPool[K, V, C]) Upsert(key K, value V) (core.UpsertState, V) {
	shard := l.mux(key)
	return l.shards[shard].Upsert(key, value)
}

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
type tPool[K comparable, V any, C core.TTLShard[K, V]] struct {
	pool[K, V, C]
	clock *clock.Clock
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
		pool: pool[K, V, C]{
			capacity: capacity,
			shards: shards,
			mux: hash,
		},
		clock: clock,
		sliding: sliding,
	}, nil
}

// Close safely closes the background clock when TTL is enabled on the cache.
func (l *tPool[K, V, C]) Close() {
	if l.clock != nil {
		l.clock.Stop()
	}
}

// Check [PoolTLRU.Get] on how Get works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *tPool[K, V, C]) GetWithTTL(key K) (V, time.Duration, bool) {
	shard := l.mux(key)
	return l.pool.shards[shard].GetWithTTL(key)
}

// Check [PoolTLRU.Peek] on how Peek works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *tPool[K, V, C]) PeekWithTTL(key K) (V, time.Duration, bool) {
	shard := l.mux(key)
	return l.pool.shards[shard].PeekWithTTL(key)
}

// PutWithTTL adds a new value to the cache with the given key and the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [PoolTLRU.Put] on how Put works.
func (l *tPool[K, V, C]) PutWithTTL(key K, value V, ttl time.Duration) {
	shard := l.mux(key)
	l.pool.shards[shard].PutWithTTL(key, value, ttl)
}

// Refresh resets the TTL of an existing key using the default ttl.
// It returns false if the key could not be found.
func (l *tPool[K, V, C]) Refresh(key K) bool {
	shard := l.mux(key)
	return l.pool.shards[shard].Refresh(key)
}

// SetTTL resets the TTL of an existing key using the provided ttl value.
// It returns false if the key could not be found.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (l *tPool[K, V, C]) SetTTL(key K, ttl time.Duration) bool {
	shard := l.mux(key)
	return l.pool.shards[shard].SetTTL(key, ttl)
}

// TTL returns the remaining TTL for the key.
func (l *tPool[K, V, C]) TTL(key K) (time.Duration, bool) {
	shard := l.mux(key)
	return l.pool.shards[shard].TTL(key)
}

// UpsertWithTTL adds a new value to the cache with the given key and the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [PoolTLRU.Upsert] on how Upsert works.
func (l *tPool[K, V, C]) UpsertWithTTL(key K, value V, ttl time.Duration) (core.UpsertState, V) {
	shard := l.mux(key)
	return l.pool.shards[shard].UpsertWithTTL(key, value, ttl)
}
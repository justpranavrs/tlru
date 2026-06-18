// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru

import (
	"cmp"
	"errors"
	"slices"

	"github.com/justpranavrs/tlru/internal/conv"
	"github.com/justpranavrs/tlru/internal/errs"
	"github.com/justpranavrs/tlru/internal/mux"
	"github.com/justpranavrs/tlru/lrucore"
)

// DefaultShards represents the number of shards allocated to LRU
// if [WithShards] option is not configured.
const DefaultShards int = 128

// Cache defines the general implementation of a 'Least Recently Used' cache.
// It has thread-safe operations.
//
// As Items are Added to the Cache, The 'Least Recently Used' key
// is evicted from the Cache.
//
// K represents the type of the key, whereas V represents the type of the Value
// in the cache
type Cache[K comparable, V any] interface {
	// Capacity returns the maximum allocated capacity of the LRU cache.
	Capacity() int

	// Compaction is an expensive O(N) operation to deal with memory fragmentation.
	Compaction()

	// Contains verifies if the key is present in the LRU Cache.
	// It does not update the key to recent status.
	Contains(key K) bool

	// Flush clears the LRU cache of all its keys and values.
	Flush()

	// Get retrieves the cache value using key.
	// It returns false if the key is not found.
	Get(key K) (V, bool)

	// GetQuiet retrieves the cache value without updating it
	// to be the most recently used.
	// It returns false if the key is not found.
	GetQuiet(key K) (V, bool)

	// Put adds a new value to the cache with the given key.
	Put(key K, value V)

	// Size returns the current size of the LRU cache.
	Size() int
}

// LRU is the better implementation of [lrucore.LRUCore]. It is a
// 'Least Recently Used' cache with many instances of [lrucore.LRUCore]
// to prevent mutual extension locks on a single instance.
// It has thread-safe but not a completely lock free operations.
//
// It doesn't work on the standard principle of LRU, rather when the mux
// routes it to a [lrucore.LRUCore] instance, it removes its 'LRU',
// but not the globally oldest item in the cache.
//
// [mux.Mux32] takes care of routing the shards to their containers consistently
// using the FNV-1a non cryptographic hashing algorithm. It uses a custom offset
// unlike the fixed offset to prevent Hash DOS attacks.
type LRU[K comparable, V any] struct {
	// capacity represents the maximum allocated space for the LRU cache.
	capacity int

	// cache is an array of the multiple instances of [lrucore.LRUCore].
	cache []*lrucore.LRUCore[K, V]

	// mux32 is the router for the shards in the cache array.
	mux32 mux.Mux32[K]

	// unsafe is the determination of speed of Mux32.
	// For more details, Refer [mux.Mux32.unsafe]
	unsafe bool
}

// LRUOption allows the use of custom options on the New method of [LRU].
type LRUOption[K comparable, V any] struct {
	// priority represent the priority order of the options to
	// be executed. It will execute in the order of lower priority
	// to higher priority.
	priority int

	// opt is the function on the [LRU] to configure its default values.
	opt func(*LRU[K, V]) error
}

// New creates a [LRU] instance with the given capacity and options. It creates
// the required [lrucore.LRUCore] instances, initiates the [mux.Mux32] for shard routing.
func New[K comparable, V any](capacity int, opts ...LRUOption[K, V]) (*LRU[K, V], error) {
	cache := make([]*lrucore.LRUCore[K, V], DefaultShards)
	lru := &LRU[K, V]{
		capacity: capacity,
		cache:    cache,
		unsafe:   false,
	}

	// sort options by priority
	slices.SortFunc(opts, func(a, b LRUOption[K, V]) int {
		return cmp.Compare(a.priority, b.priority)
	})

	for _, opt := range opts { // options
		if err := opt.opt(lru); err != nil {
			return nil, err
		}
	}

	// initialize the mux
	lru.mux32 = mux.New32[K](len(lru.cache), lru.unsafe)

	num := len(lru.cache)
	cap := capacity / num // create lrucore instances
	rem := capacity % num
	for i := range rem {
		c, err := lrucore.New[K, V](1 + cap)
		if err != nil {
			if errors.Is(err, errs.ErrCoreInvalidCapacity) {
				return nil, errs.ErrInvalidCapacity
			}
			return nil, err
		}
		lru.cache[i] = c
	}
	for i := rem; i < num; i++ {
		c, err := lrucore.New[K, V](cap)
		if err != nil {
			return nil, err
		}
		lru.cache[i] = c
	}
	return lru, nil
}

// WithShards assigns the [LRU] instance with num shards. Shards are separate instances
// of [lrucore.LRUCore] to prevent mutex locks from slowing down the cache.
//
// For performance optimizations, number of shards are automatically rounded up
// to the next power of 2.
//
// Returns [err.ErrNoShards] if num is less than 1.
func WithShards[K comparable, V any](num int) LRUOption[K, V] {
	cnt := conv.NextPowerOf2(num)
	return LRUOption[K, V]{
		priority: 1,
		opt: func(l *LRU[K, V]) error {
			if cnt < 1 {
				return errs.ErrNoShards
			}
			l.cache = make([]*lrucore.LRUCore[K, V], cnt)
			return nil
		},
	}
}

// WithUnsafe provides a faster routing method to [Mux32] by using
// Go's unsafe package. It does byte conversion with worrying about
// Go's Garbage Collector (GC).
func WithUnsafe[K comparable, V any]() LRUOption[K, V] {
	return LRUOption[K, V]{
		priority: 2,
		opt: func(l *LRU[K, V]) error {
			l.unsafe = true
			return nil
		},
	}
}

// Capacity returns the maximum allocated capacity of the LRU cache
// across all sharded instances of [lrucore.LRUCore].
func (l *LRU[K, V]) Capacity() int {
	return l.capacity
}

// Compaction is an expensive O(N) operation to deal with memory fragmentation.
// It compacts all keys across the sharded architecture.
// For more details, refer [lrucore.LRUCore.Compaction]
func (l *LRU[K, V]) Compaction() {
	for _, c := range l.cache {
		c.Compaction()
	}
}

// Contains verifies if the key is present in the LRU Cache
// by checking the correct sharded instance.
// It does not update the key to recent status.
func (l *LRU[K, V]) Contains(key K) bool {
	shard, ok := l.mux32.Get(key)
	if !ok {
		return false
	}
	return l.cache[shard].Contains(key)
}

// Flush clears the LRU cache of all its keys and values across
// all sharded instances.
func (l *LRU[K, V]) Flush() {
	for _, c := range l.cache {
		c.Flush()
	}
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
// It updates the key as 'recent' only in its respective shard.
func (l *LRU[K, V]) Get(key K) (V, bool) {
	shard, ok := l.mux32.Get(key)
	if !ok {
		return *new(V), false
	}

	value, ok := l.cache[shard].Get(key)
	if !ok {
		return *new(V), false
	}
	return value, true
}

// GetQuiet retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found.
func (l *LRU[K, V]) GetQuiet(key K) (V, bool) {
	shard, ok := l.mux32.Get(key)
	if !ok {
		return *new(V), false
	}

	value, ok := l.cache[shard].GetQuiet(key)
	if !ok {
		return *new(V), false
	}
	return value, true
}

// Put adds a new value to the cache with the given key.
// It updates the key as 'recent' only in its respective shard.
// It evicts the key only from the respective shard the key is linked to.
func (l *LRU[K, V]) Put(key K, value V) {
	shard, _ := l.mux32.Get(key)
	l.cache[shard].Put(key, value)
}

// Size returns the current size of the LRU cache
// across all sharded instances.
func (l *LRU[K, V]) Size() int {
	size := 0
	for _, c := range l.cache {
		size += c.Size()
	}
	return size
}

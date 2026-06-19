// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru

import (
	"errors"
	"math"
	"sync/atomic"

	"github.com/justpranavrs/tlru/internal/errs"
	"github.com/justpranavrs/tlru/internal/mathutil"
	"github.com/justpranavrs/tlru/lrucore"
	"github.com/justpranavrs/tlru/mux"
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

	// Peek retrieves the cache value without updating it
	// to be the most recently used.
	// It returns false if the key is not found.
	Peek(key K) (V, bool)

	// Put adds a new value to the cache with the given key.
	Put(key K, value V)

	// PutGrew adds a new value to the cache with the given key.
	// It returns true if the size of the cache has grown, else returns false.
	PutGrew(key K, value V) bool

	// Size returns the current size of the LRU cache.
	Size() int
}

// LRU is the better implementation of [lrucore.LRUCore]. It is a
// 'Least Recently Used' cache with many instances of [lrucore.LRUCore]
// to prevent mutual extension locks on a single instance.
// It has thread-safe but not a completely lock free operations.
//
// It doesn't work on the standard principle of LRU, rather when the mux
// routes it to a [lrucore.LRUCore] instance. LRU works on shard-local based eviction
// not on globally oldest item in the cache.
//
// [mux.Mux] takes care of routing the shards to their containers consistently
// using its hashing algorithm. The default mux is [mux.MuxX32.Get].
type LRU[K comparable, V any] struct {
	// capacity represents the maximum allocated space for the LRU cache.
	capacity int

	// cache is an array of the multiple instances of [lrucore.LRUCore].
	cache []*lrucore.LRUCore[K, V]

	// mux is the router for the shards in the cache array.
	// It takes in a key K and outputs a hash [uint32]
	mux mux.Mux[K]

	// size measures the current allocated space of the cache.
	// It uses atomic.Int32 to monitor without data races.
	size atomic.Int32
}

// config represents the configuration of [LRU]. It should be used with [Option].
type config struct {
	shards int
	mux    any
}

// Option is used to configure [LRU] when creating an instance using [New] constructor.
type Option func(c *config) error

// New creates a [LRU] instance with the given capacity and options. It creates
// the required [lrucore.LRUCore] instances, initiates the [mux.Mux] for shard routing.
//
// Returns [errs.ErrInvalidShards] if shards is not greater than 0 and in [int32] range.
//
// Returns [errs.ErrInvalidCapacity] if capacity is not in [int32] range and greater than number
// of shards.
func New[K comparable, V any](capacity int, opts ...Option) (*LRU[K, V], error) {
	// build the config
	cfg := config{
		shards: DefaultShards,
		mux:    nil,
	}
	for _, opt := range opts { // options
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	if capacity <= int(cfg.shards) || (capacity > math.MaxInt32-1) {
		return nil, errs.ErrInvalidCapacity
	}

	// set the mux hash
	var hash mux.Mux[K]
	if cfg.mux != nil {
		hash = cfg.mux.(mux.Mux[K])
	} else {
		fun, err := mux.NewX32[K](cfg.shards)
		if err != nil {
			return nil, err
		}
		hash = fun
	}

	lru := &LRU[K, V]{
		capacity: capacity,
		cache:    make([]*lrucore.LRUCore[K, V], cfg.shards),
		mux:      hash,
	}

	cap := capacity / cfg.shards // create lrucore instances
	rem := capacity & (cfg.shards - 1)
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
	for i := rem; i < cfg.shards; i++ {
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
// Returns [errs.ErrInvalidShards] if num is not in [int32] range or equals zero.
func WithShards(num int) Option {
	cnt := mathutil.NextPowerOf2(num)
	return func(c *config) error {
		if cnt < 1 || (cnt > 1<<32-1) {
			return errs.ErrInvalidShards
		}
		c.shards = cnt
		return nil
	}
}

// WithMux requires a custom [mux.Mux] type function. It is used
// with [LRU] to configure its mux, which is responsible for routing the shards.
func WithMux[K comparable](cm mux.Mux[K]) Option {
	return func(c *config) error {
		c.mux = cm
		return nil
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
	shard := l.mux(key)
	return l.cache[shard].Contains(key)
}

// Flush clears the LRU cache of all its keys and values across
// all sharded instances.
func (l *LRU[K, V]) Flush() {
	for _, c := range l.cache {
		c.Flush()
	}
	l.size.Store(0)
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
// It updates the key as 'recent' only in its respective shard.
func (l *LRU[K, V]) Get(key K) (V, bool) {
	shard := l.mux(key)
	value, ok := l.cache[shard].Get(key)
	if !ok {
		return *new(V), false
	}
	return value, true
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found.
func (l *LRU[K, V]) Peek(key K) (V, bool) {
	shard := l.mux(key)
	value, ok := l.cache[shard].Peek(key)
	if !ok {
		return *new(V), false
	}
	return value, true
}

// Put adds a new value to the cache with the given key.
// It updates the key as 'recent' only in its respective shard.
// It evicts the key only from the respective shard the key is linked to.
func (l *LRU[K, V]) Put(key K, value V) {
	l.PutGrew(key, value)
}

// PutGrew adds a new value to the cache with the given key.
// It returns true if the size of the cache has grown, else returns false.
// It evicts or updates locally on the shard, instead of global cache.
func (l *LRU[K, V]) PutGrew(key K, value V) bool {
	shard := l.mux(key)
	if l.cache[shard].PutGrew(key, value) {
		l.size.Add(1)
		return true
	}
	return false
}

// Shards returns the number of sharded instances in the LRU cache.
func (l *LRU[K, V]) Shards() int {
	return len(l.cache)
}

// Size returns the current size of the LRU cache
// across all sharded instances.
func (l *LRU[K, V]) Size() int {
	return int(l.size.Load())
}

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

// DefaultShards represents the number of shards allocated to LRU
// if [WithShards] option is not configured.
const DefaultShards int = 128

// Cache defines the general implementation of a 'Least Recently Used' cache.
// It has thread-safe operations.
//
// K represents the type of the key, whereas V represents the type of the Value
// in the cache
type Cache[K comparable, V any] interface {
	// Capacity returns the maximum allocated capacity of the LRU cache.
	Capacity() int

	// Close safely closes the background clock when TTL is enabled on the cache.
	Close()

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

// LRU is the better implementation of [lrucore.Shard]. It is a
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
type LRU[K comparable, V any, C lrucore.Shard[K, V]] struct {
	// capacity represents the maximum allocated space for the LRU cache.
	capacity int

	// cache is an array of the multiple instances of [lrucore.Core].
	cache []C

	// mux is the router for the shards in the cache array.
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
func New[K comparable, V any](capacity int, opts ...LRUOption) (*LRU[K, V, *lrucore.Core[K, V]], error) {
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

	return initLRU(capacity, cfg.shards, hash, func(cap int) (*lrucore.Core[K, V], error) {
		return lrucore.New[K, V](cap)
	})
}

// initLRU creates a [LRU] instance with the provided arguments. It creates shards based on the createShard function.
func initLRU[K comparable, V any, C lrucore.Shard[K, V]](capacity int, shards int, hash mux.Mux[K], createShard func(cap int) (C, error)) (*LRU[K, V, C], error) {
	if capacity < int(shards*2) || (capacity > math.MaxInt32) {
		return nil, ErrInvalidCapacity
	}

	var err error
	cache := make([]C, shards)

	cap := capacity / shards // create lrucore instances
	rem := capacity % shards
	for i := range shards {
		sCap := cap
		if i < rem {
			sCap++
		}

		cache[i], err = createShard(sCap)
		if err != nil {
			if errors.Is(err, lrucore.ErrInvalidCapacity) {
				return nil, ErrInvalidCapacity
			}
			return nil, err
		}
	}
	return &LRU[K, V, C]{
		capacity: capacity,
		cache:    cache,
		mux:      hash,
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
// across all sharded instances of [lrucore.Core].
func (l *LRU[K, V, C]) Capacity() int {
	return l.capacity
}

// Close safely closes the all the underlying shards with its Close().
func (l *LRU[K, V, C]) Close() {
	for _, c := range l.cache {
		c.Close()
	}
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *LRU[K, V, C]) Delete(key K) (V, bool) {
	shard := l.mux(key)
	value, ok := l.cache[shard].Delete(key)
	if !ok {
		return *new(V), false
	}
	return value, true
}

// Flush clears the LRU cache of all its keys and values across
// all sharded instances.
func (l *LRU[K, V, C]) Flush() {
	for _, c := range l.cache {
		c.Flush()
	}
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
// It updates the key as 'recent' only in its respective shard.
func (l *LRU[K, V, C]) Get(key K) (V, bool) {
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
func (l *LRU[K, V, C]) Peek(key K) (V, bool) {
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
func (l *LRU[K, V, C]) Put(key K, value V) {
	l.Upsert(key, value)
}

// ResetStats resets the stats of the sharded LRU cache.
func (l *LRU[K, V, C]) ResetStats() {
	for i := range l.cache {
		l.cache[i].ResetStats()
	}
}

// Shards returns the number of sharded instances in the LRU cache.
func (l *LRU[K, V, C]) Shards() int {
	return len(l.cache)
}

// Size returns the current size of the LRU cache
// across all sharded instances.
func (l *LRU[K, V, C]) Size() int {
	size := 0
	for i := range l.cache {
		size += l.cache[i].Size()
	}
	return size
}

// Stats return the current stats of the sharded LRU cache.
func (l *LRU[K, V, C]) Stats() lrucore.CoreStats {
	stats := lrucore.CoreStats{}
	for i := range l.cache {
		st := l.cache[i].Stats()
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
func (l *LRU[K, V, C]) Upsert(key K, value V) (lrucore.UpsertState, V) {
	shard := l.mux(key)
	return l.cache[shard].Upsert(key, value)
}

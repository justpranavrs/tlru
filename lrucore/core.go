// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lrucore

import (
	"errors"
	"math"
	"sync"
)

// Shard defines the most basic blueprint of a 'Least Recently Used' cache.
// It has thread-safe operations.
//
// As Items are Added to the Cache, The 'Least Recently Used' key
// is evicted from the Cache.
//
// K represents the type of the key, whereas V represents the type of the Value
// in the cache
type Shard[K comparable, V any] interface {
	// Capacity returns the maximum allocated capacity of the LRU cache.
	Capacity() int

	// Contains checks whether the key is present in the Cache.
	Contains(key K) bool

	// Delete removes the key from the cache and returns the evicted value.
	// It returns false if the key was not found in the cache.
	Delete(key K) (V, bool)

	// Flush clears the LRU cache of all its keys and values.
	Flush()

	// Get retrieves the cache value using key.
	// It returns false if the key is not found.
	Get(key K) (V, bool)

	// GetMany allows retrieval of multiple keys at the same time under a single internal lock.
	GetMany(keys []K, values []V, exists []bool) error

	// Peek retrieves the cache value without updating it
	// to be the most recently used.
	// It returns false if the key is not found.
	Peek(key K) (V, bool)

	// Put adds a new value to the cache with the given key.
	Put(key K, value V)

	// PutMany adds multiple key-value pairs at the same time under a single internal lock.
	PutMany(keys []K, values []V) error

	// ResetStats resets the stats of the LRU cache.
	ResetStats()

	// Size returns the current size of the LRU cache.
	Size() int

	// Stats return the current stats of the LRU cache.
	Stats() CoreStats

	// Upsert adds a new value to the cache with the given key.
	// It returns a value based on how the internal state of the cache changed.
	Upsert(key K, value V) (UpsertState, V)
}

// Core is basic implementation of 'Least Recently Used' Cache.
// It uses a contiguous array of nodes as opposed to the standard doubly-linked list.
// It uses a hash map to track the key to the index in the array.
// It has thread-safe operations.
//
// It is recommended for key K to be primitives such as ([int], [uint64], [string]).
type Core[K comparable, V any] struct {
	// mu is a mutual extension lock.
	mu sync.Mutex

	// hash maps the key to a unique index in the nodes array.
	hash map[K]int32

	// links are the connections of the doubly-linked list.
	links []link

	// nodes are the elements of the doubly-linked list.
	// It is an array of [node].
	nodes []node[K, V]

	// mru points to the most recently used key in the cache.
	// its next link will be the next free slot.
	mru int32

	// tail represents the last index in the nodes array.
	// It has a value of capacity + 1.
	tail int32

	// capacity represents the maximum allocated space for the cache.
	capacity int

	// stats measures the instance's hits, misses and evictions.
	stats CoreStats
}

// node represents a single element in [Core] instance's internal
// doubly-linked list. It stores the data. Refer [link] for the
// linked list implementation.
type node[K comparable, V any] struct {
	// key is the identifier used to lookup in the [Core] cache.
	// It is recommended for key to be primitives such as (int, uint64, string).
	key K

	// value holds the actual data stored in the cache
	value V
}

// link represents a node's connections in the doubly-linked list.
// It is packed into a contiguous array to maximize CPU cache
// and avoid memory fragmentation.
type link struct {
	// prev holds the index to the previous element in the doubly-linked list.
	prev int32

	// next holds the index to the next element in the doubly-linked list.
	next int32
}

// CoreStats represents the metrics of a [Core] instance.
type CoreStats struct {
	// Hits is the number of successful cache lookups from [Core.Get] and [Core.GetMany].
	Hits int

	// Misses is the number of failed cache lookups from [Core.Get] and [Core.GetMany].
	Misses int

	// Evictions is the number of keys removed from cache during [Core.Put] and [Core.PutMany].
	Evictions int

	// Expirations is the number of expirations triggered due to TTL (Time-To-Live).
	Expirations int
}

// UpsertState represents the value returned by Upsert operation.
type UpsertState uint8

const (
	// AddAfterExpiration is triggered when a new key was added after an expiration
	// due to TTL (Time-To-Live).
	AddAfterExpiration UpsertState = iota

	// AddNoEvict is triggered when a new key was added without eviction.
	AddNoEvict

	// AddOnEvict is triggered when a new key was added after an eviction.
	AddOnEvict

	// Replace is triggered when an older key's value was replaced.
	Replace
)

var (
	// ErrInvalidBatchSize is returned by batch operations when keys and values do not have the same lengths.
	ErrInvalidBatchSize = errors.New("invalid LRU batch sizes: keys and values do not have the same lengths")

	// ErrInvalidCapacity is returned by [New] or [NewTTL] when the maximum cache capacity is not in [2, 2147483645].
	ErrInvalidCapacity = errors.New("invalid LRU cache capacity: must be in the range [2, 2147483645]")
)

// New creates an instance of [Core] using the given capacity.
//
// Returns an [ErrInvalidCapacity] if the capacity is not in [2, 2147483645].
func New[K comparable, V any](capacity int) (*Core[K, V], error) {
	if capacity < 2 || (capacity > math.MaxInt32-2) { // For capacity 1, a variable can be used.
		return nil, ErrInvalidCapacity
	}
	cap := int32(capacity)
	tail := 1 + cap

	// allocate the initial nodes and links array with a size of 2+cap
	nodes := make([]node[K, V], 2+capacity)
	links := make([]link, 2+capacity)

	links[0].next = 1
	links[tail].prev = cap

	i := int32(1)
	for ; i <= cap; i++ {
		links[i].prev = i - 1
		links[i].next = i + 1
	}

	// lru
	return &Core[K, V]{
		hash:     make(map[K]int32, capacity),
		nodes:    nodes,
		links:    links,
		capacity: capacity,
		tail:     tail,
		mru:      0,
		stats: CoreStats{
			Hits:        0,
			Misses:      0,
			Evictions:   0,
			Expirations: 0,
		},
	}, nil
}

// Capacity returns the maximum allocated capacity of the LRU cache.
func (l *Core[K, V]) Capacity() int {
	return l.capacity
}

// Contains checks whether the key is present in the Cache.
func (l *Core[K, V]) Contains(key K) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.peekKey(key)
	return ok
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *Core[K, V]) Delete(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	curr, ok := l.hash[key]
	if !ok {
		return *new(V), false
	}
	l.stats.Evictions++
	return l.deleteKey(curr), true
}

// Flush clears the LRU cache of all its keys and values.
func (l *Core[K, V]) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()

	clear(l.hash)
	clear(l.nodes)
	l.mru = 0
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
func (l *Core[K, V]) Get(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.getKey(key)
}

// GetMany allows retrieval of multiple keys at the same time under a single internal lock.
//
// The keys, values and exists array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
//
// The operation modifies values and exists in-place. If a key is not present in the cache,
// the corresponding index in exists is set to false and leaves the value at that index unchanged.
func (l *Core[K, V]) GetMany(keys []K, values []V, exists []bool) error {
	if (len(keys) != len(values)) || (len(keys) != len(exists)) {
		return ErrInvalidBatchSize
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := range keys {
		val, ok := l.getKey(keys[i])
		exists[i] = ok
		if ok {
			values[i] = val
		}
	}
	return nil
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found.
func (l *Core[K, V]) Peek(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.peekKey(key)
}

// Put adds a new value to the cache with the given key.
// See [Core.Upsert] for detailed information on cache state transitions.
func (l *Core[K, V]) Put(key K, value V) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.putKey(key, value)
}

// PutMany allows the addition of multiple key-value pairs at the
// same time under a single internal lock.
//
// The keys and values array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
func (l *Core[K, V]) PutMany(keys []K, values []V) error {
	if len(keys) != len(values) {
		return ErrInvalidBatchSize
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := range keys {
		l.putKey(keys[i], values[i])
	}
	return nil
}

// ResetStats resets the stats of the LRU cache.
func (l *Core[K, V]) ResetStats() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.stats = CoreStats{}
}

// Size returns the current size of the LRU cache.
func (l *Core[K, V]) Size() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.hash)
}

// Stats return the current stats of the LRU cache.
func (l *Core[K, V]) Stats() CoreStats {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.stats
}

// Upsert adds a new value to the cache with the given key.
// It returns [UpsertState] based on how the internal state of the cache changed.
//
// It also returns a value based on [UpsertState]
//   - [AddNoEvict] returns the zero value of V.
//   - [AddOnEvict] returns the evicted value.
//   - [Replace] returns the old value the key had.
func (l *Core[K, V]) Upsert(key K, value V) (UpsertState, V) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.putKey(key, value)
}

// addKey adds the cache with the given key and cache value.
func (l *Core[K, V]) addKey(key K, value V) {
	free := l.links[l.mru].next
	l.nodes[free] = node[K, V]{
		key:   key,
		value: value,
	}

	l.hash[key] = free
	l.mru = free
}

// getIndex updates the recency, the stats and returns value with the index.
func (l *Core[K, V]) getIndex(idx int32) V {
	if idx != l.mru { // not already recent
		l.makeRecent(idx)
	}
	l.stats.Hits++
	val := l.nodes[idx].value
	return val
}

// getKey retrieves the value of the key passed as argument.
// It returns false if the key is not present.
func (l *Core[K, V]) getKey(key K) (V, bool) {
	curr, ok := l.hash[key]
	if !ok {
		l.stats.Misses++
		return *new(V), false // not present in cache
	}
	return l.getIndex(curr), true
}

// makeRecent sets the key as 'Most Recently Used'.
func (l *Core[K, V]) makeRecent(idx int32) {
	l.remove(idx)

	free := l.links[l.mru].next
	l.links[l.mru].next = idx
	l.links[free].prev = idx

	l.links[idx] = link{
		prev: l.mru,
		next: free,
	}
	l.mru = idx
}

// peekKey retrieves value of key without updating cache internal state.
// Returns false when key not found.
func (l *Core[K, V]) peekKey(key K) (V, bool) {
	curr, ok := l.hash[key]
	if !ok {
		return *new(V), false
	}
	return l.nodes[curr].value, true
}

// putKey adds or updates the cache value with key "key".
// It also returns a value based on whether it was evicted or replaced.
func (l *Core[K, V]) putKey(key K, value V) (UpsertState, V) {
	curr, ok := l.hash[key]
	if !ok { // not present in cache
		if len(l.hash) == l.capacity {
			old := l.links[0].next
			oldVal := l.nodes[old].value

			l.deleteKey(old)
			l.stats.Evictions++

			l.addKey(key, value)
			return AddOnEvict, oldVal
		} else {
			l.addKey(key, value)
			return AddNoEvict, *new(V)
		}
	} else { // present in cache, just update values
		oldVal := l.nodes[curr].value

		l.nodes[curr].value = value
		if curr != l.mru { // not already recent
			l.makeRecent(curr)
		}
		return Replace, oldVal
	}
}

// remove detaches the given element from the doubly-linked list.
func (l *Core[K, V]) remove(idx int32) {
	curr := l.links[idx]

	l.links[curr.prev].next = curr.next
	l.links[curr.next].prev = curr.prev
}

// deleteKey removes a key from the cache.
func (l *Core[K, V]) deleteKey(idx int32) V {
	l.remove(idx)

	free := l.links[l.tail].prev

	l.links[l.tail].prev = idx
	l.links[free].next = idx
	l.links[idx] = link{
		prev: free,
		next: l.tail,
	}

	val := l.nodes[idx].value
	delete(l.hash, l.nodes[idx].key)

	l.nodes[idx] = node[K, V]{}
	return val
}

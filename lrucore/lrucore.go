// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lrucore

import (
	"errors"
	"math"
	"sync"
)

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

	// tail represents the last index in the nodes array.
	// It has a value of capacity + 1.
	tail int32

	// free points to the next available space for the cache.
	// it sits on the last occupied key and right before insertion,
	// it changes to its new insertion index.
	free int32

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
}

// UpsertState represents the value returned by Upsert operation.
type UpsertState uint8

const (
	// AddNoEvict is triggered when a new key was added without eviction.
	AddNoEvict UpsertState = iota

	// AddOnEvict is triggered when a new key was added after an eviction.
	AddOnEvict

	// Replace is triggered when an older key's value was replaced.
	Replace
)

var (
	// ErrInvalidBatchSize is returned by [Core.GetMany] or [Core.PutMany] when keys and values do not have the same lengths.
	ErrInvalidBatchSize = errors.New("invalid LRU batch sizes: keys and values do not have the same lengths")

	// ErrInvalidCapacity is returned by [New] when the maximum cache capacity is not in [2, 2147483646].
	ErrInvalidCapacity = errors.New("invalid LRU cache capacity: must be in the range [2, 2147483646]")
)

// New creates an instance of [Core] using the given capacity.
//
// Returns an [ErrInvalidCapacity] if the capacity is not in [2, 2147483646].
func New[K comparable, V any](capacity int) (*Core[K, V], error) {
	if capacity < 2 || (capacity > math.MaxInt32-2) { // For capacity 1, a variable can be used.
		return nil, ErrInvalidCapacity
	}
	tail := 1 + int32(capacity)

	// allocate the initial nodes and links array with a size of 2+cap
	nodes := make([]node[K, V], 2+capacity)
	links := make([]link, 2+capacity)

	links[0].next = tail
	links[tail].prev = 0

	// lru
	lru := &Core[K, V]{
		hash:     make(map[K]int32, capacity),
		nodes:    nodes,
		links:    links,
		capacity: capacity,
		tail:     tail,
		free:     0,
		stats: CoreStats{
			Hits:      0,
			Misses:    0,
			Evictions: 0,
		},
	}
	return lru, nil
}

// Capacity returns the maximum allocated capacity of the LRU cache.
func (l *Core[K, V]) Capacity() int {
	return l.capacity
}

// Flush clears the LRU cache of all its keys and values.
func (l *Core[K, V]) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()

	clear(l.hash)
	clear(l.nodes)
	clear(l.links)

	l.links[0].next = l.tail
	l.links[l.tail].prev = 0
	l.free = 0
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
// the corresponding index in exists tis set to false and leaves the value at that index unchanged.
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

// Shards returns the number of sharded instances in the LRU cache.
// For [lrucore.Core], this is always 1.
func (l *Core[K, V]) Shards() int {
	return 1
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
// It returns a value based on how the internal state of the cache changed.
func (l *Core[K, V]) Upsert(key K, value V) UpsertState {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.putKey(key, value)
}

// addKey adds the cache with the given key and cache value.
func (l *Core[K, V]) addKey(key K, value V) {
	l.nodes[l.free] = node[K, V]{
		key:   key,
		value: value,
	}
	l.links[l.free] = link{
		prev: l.links[l.tail].prev,
		next: l.tail,
	}

	l.links[l.links[l.tail].prev].next = l.free
	l.links[l.tail].prev = l.free

	l.hash[key] = l.free
}

// getKey retrieves the value of the key passed as argument.
// It returns false if the key is not present.
func (l *Core[K, V]) getKey(key K) (V, bool) {
	curr, ok := l.hash[key]
	if !ok {
		l.stats.Misses++
		return *new(V), false // not present in cache
	}
	if curr != l.links[l.tail].prev { // not already recent
		l.remove(curr)
		l.makeRecent(curr)
	}
	l.stats.Hits++
	val := l.nodes[curr].value
	return val, true
}

// makeRecent sets the key as 'Most Recently Used'.
func (l *Core[K, V]) makeRecent(idx int32) {
	l.links[l.links[l.tail].prev].next = idx
	l.links[idx].prev = l.links[l.tail].prev

	l.links[idx].next = l.tail
	l.links[l.tail].prev = idx
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
func (l *Core[K, V]) putKey(key K, value V) UpsertState {
	curr, ok := l.hash[key]
	if !ok { // not present in cache
		if len(l.hash) == l.capacity {
			l.removeOld()
			l.addKey(key, value)
			return AddOnEvict
		} else {
			l.free++
			l.addKey(key, value)
			return AddNoEvict
		}
	} else { // present in cache, just update values
		l.nodes[curr].value = value
		if curr != l.links[l.tail].prev { // not already recent
			l.remove(curr)
			l.makeRecent(curr)
		}
		return Replace
	}
}

// remove detaches the given element from the doubly-linked list.
func (l *Core[K, V]) remove(idx int32) {
	curr := l.links[idx]

	l.links[curr.prev].next = curr.next
	l.links[curr.next].prev = curr.prev
}

// removeOld removes the 'Least Recently Used' cache.
func (l *Core[K, V]) removeOld() {
	old := l.links[0].next
	l.remove(old)

	delete(l.hash, l.nodes[old].key)

	l.free = old // set the next free index to the dropped cache value
	l.stats.Evictions++
}

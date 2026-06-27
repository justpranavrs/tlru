// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import "math"

// base is the internal blueprint for all LRU which will determine
// the core architecture.
type base[K comparable, V any] interface {
	Shard[K, V]

	// addEvictions increases the evictions stat by 1.
	addEvictions()

	// addExpirations increases the expirations stat by 1.
	addExpirations()

	// addHits increases the hits stat by 1.
	addHits()

	// addKey adds the cache with the given key and cache value.
	addKey(key K, value V)

	// addMisses increases the misses stat by 1.
	addMisses()

	// clearState erases the entire cache's internal state
	clearState()

	// deleteWithIndex removes a node from the cache.
	deleteWithIndex(idx int32) node[K, V]

	// deleteWithKey removes a key from the cache.
	deleteWithKey(key K) (V, bool)

	// getWithIndex updates the recency, the stats and returns the value.
	getWithIndex(idx int32) V

	// getWithKey retrieves the value of the key passed as argument.
	// It returns false if the key is not present.
	getWithKey(key K) (V, bool)

	// makeRecent sets the key as 'Most Recently Used'.
	makeRecent(idx int32)

	// peekWithIndex retrieves the node without updating recency.
	peekWithIndex(idx int32) node[K, V]

	// peekWithKey retrieves value of key without updating cache internal state.
	// Returns false when key not found.
	peekWithKey(key K) (V, bool)

	// putWithKey adds or updates the cache value with key "key".
	// It also returns a value based on whether it was evicted or replaced.
	putWithKey(key K, value V) (UpsertState, V)

	// removeOldest evicts the oldest item in the cache.
	removeOldest() node[K, V]

	// removeWithIndex detaches the given element from the doubly-linked list.
	removeWithIndex(idx int32)

	// retrieveIndexWithKey retrieves the index from the hash map.
	retrieveIndexWithKey(key K) (int32, bool)

	// updateWithIndex updates the value of the provided index.
	updateWithIndex(idx int32, value V)
}

// lruBase is the initial unlocked doubly-linked list implementation of
// Least Recently Used cache. It is not safe for concurrent workloads.
// [syncBase] provides the [sync.Mutex] locks for it.
type lruBase[K comparable, V any] struct {
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
	stats Stats
}

// node represents a single element in [lruBase] instance's internal
// doubly-linked list. It stores the data. Refer [link] for the
// linked list implementation.
type node[K comparable, V any] struct {
	// key is the identifier used to lookup in the [lruBase] cache.
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

// assembleLRU creates an instance of [lruBase] using the given capacity.
func assembleLRU[K comparable, V any](capacity int) (*lruBase[K, V], error) {
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
	return &lruBase[K, V]{
		hash:     make(map[K]int32, capacity),
		nodes:    nodes,
		links:    links,
		capacity: capacity,
		tail:     tail,
		mru:      0,
	}, nil
}

// Capacity returns the maximum allocated capacity of the LRU cache.
func (l *lruBase[K, V]) Capacity() int {
	return l.capacity
}

// Contains checks whether the key is present in the Cache.
func (l *lruBase[K, V]) Contains(key K) bool {
	_, ok := l.peekWithKey(key)
	return ok
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *lruBase[K, V]) Delete(key K) (V, bool) {
	return l.deleteWithKey(key)
}

// Flush clears the LRU cache of all its keys and values.
func (l *lruBase[K, V]) Flush() {
	l.clearState()
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
func (l *lruBase[K, V]) Get(key K) (V, bool) {
	return l.getWithKey(key)
}

// GetMany allows retrieval of multiple keys at the same time.
//
// The keys, values and exists array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
//
// The operation modifies values and exists in-place. If a key is not present in the cache,
// the corresponding index in exists is set to false and leaves the value at that index unchanged.
func (l *lruBase[K, V]) GetMany(keys []K, values []V, exists []bool) error {
	if (len(keys) != len(values)) || (len(keys) != len(exists)) {
		return ErrInvalidBatchSize
	}

	for i := range keys {
		val, ok := l.getWithKey(keys[i])
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
func (l *lruBase[K, V]) Peek(key K) (V, bool) {
	return l.peekWithKey(key)
}

// Put adds a new value to the cache with the given key.
// See [LRU.Upsert] for detailed information on cache state transitions.
func (l *lruBase[K, V]) Put(key K, value V) {
	l.putWithKey(key, value)
}

// PutMany allows the addition of multiple key-value pairs at the
// same time.
//
// The keys and values array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
func (l *lruBase[K, V]) PutMany(keys []K, values []V) error {
	if len(keys) != len(values) {
		return ErrInvalidBatchSize
	}

	for i := range keys {
		l.putWithKey(keys[i], values[i])
	}
	return nil
}

// ResetStats resets the stats of the LRU cache.
func (l *lruBase[K, V]) ResetStats() {
	l.stats = Stats{}
}

// Size returns the current size of the LRU cache.
func (l *lruBase[K, V]) Size() int {
	return len(l.hash)
}

// Stats return the current stats of the LRU cache.
func (l *lruBase[K, V]) Stats() Stats {
	return l.stats
}

// Upsert adds a new value to the cache with the given key.
// It returns [UpsertState] based on how the internal state of the cache changed.
//
// It also returns a value based on [UpsertState]
//   - [UpsertAddNoEviction] returns the zero value of V.
//   - [UpsertAddWithEviction] returns the evicted value.
//   - [UpsertReplace] returns the old value the key had.
func (l *lruBase[K, V]) Upsert(key K, value V) (UpsertState, V) {
	return l.putWithKey(key, value)
}

// addEvictions increases the evictions stat by 1.
func (l *lruBase[K, V]) addEvictions() {
	l.stats.Evictions++
}

// addExpirations increases the expirations stat by 1.
func (l *lruBase[K, V]) addExpirations() {
	l.stats.Expirations++
}

// addHits increases the hits stat by 1.
func (l *lruBase[K, V]) addHits() {
	l.stats.Hits++
}

// addKey adds the cache with the given key and cache value.
func (l *lruBase[K, V]) addKey(key K, value V) {
	free := l.links[l.mru].next
	l.nodes[free] = node[K, V]{
		key:   key,
		value: value,
	}

	l.hash[key] = free
	l.mru = free
}

// addMisses increases the misses stat by 1.
func (l *lruBase[K, V]) addMisses() {
	l.stats.Misses++
}

// clearState erases the entire cache's internal state
func (l *lruBase[K, V]) clearState() {
	clear(l.hash)
	clear(l.nodes)
	l.mru = 0
}

// deleteWithIndex removes a node from the cache.
func (l *lruBase[K, V]) deleteWithIndex(idx int32) node[K, V] {
	l.removeWithIndex(idx)

	free := l.links[l.tail].prev

	l.links[l.tail].prev = idx
	l.links[free].next = idx
	l.links[idx] = link{
		prev: free,
		next: l.tail,
	}

	kv := l.nodes[idx]
	delete(l.hash, l.nodes[idx].key)

	l.nodes[idx] = node[K, V]{}
	return kv
}

// deleteWithKey removes a key from the cache.
func (l *lruBase[K, V]) deleteWithKey(key K) (V, bool) {
	curr, ok := l.hash[key]
	if !ok {
		return *new(V), false
	}
	return l.deleteWithIndex(curr).value, true
}

// getWithIndex updates the recency, the stats and returns the value.
func (l *lruBase[K, V]) getWithIndex(idx int32) V {
	l.makeRecent(idx)
	l.addHits()
	return l.peekWithIndex(idx).value
}

// getWithKey retrieves the value of the key passed as argument.
// It returns false if the key is not present.
func (l *lruBase[K, V]) getWithKey(key K) (V, bool) {
	curr, ok := l.hash[key]
	if !ok {
		l.addMisses()
		return *new(V), false // not present in cache
	}
	return l.getWithIndex(curr), true
}

// makeRecent sets the key as 'Most Recently Used'.
func (l *lruBase[K, V]) makeRecent(idx int32) {
	if idx == l.mru { // already recent
		return
	}
	l.removeWithIndex(idx)

	free := l.links[l.mru].next
	l.links[l.mru].next = idx
	l.links[free].prev = idx

	l.links[idx] = link{
		prev: l.mru,
		next: free,
	}
	l.mru = idx
}

// peekWithIndex retrieves the node without updating recency.
func (l *lruBase[K, V]) peekWithIndex(idx int32) node[K, V] {
	return l.nodes[idx]
}

// peekWithKey retrieves value of key without updating cache internal state.
// Returns false when key not found.
func (l *lruBase[K, V]) peekWithKey(key K) (V, bool) {
	curr, ok := l.hash[key]
	if !ok {
		return *new(V), false
	}
	return l.peekWithIndex(curr).value, true
}

// putWithKey adds or updates the cache value with key "key".
// It also returns a value based on whether it was evicted or replaced.
func (l *lruBase[K, V]) putWithKey(key K, value V) (UpsertState, V) {
	curr, ok := l.hash[key]
	if !ok { // not present in cache
		if len(l.hash) == l.capacity {
			kv := l.removeOldest()

			l.addKey(key, value)
			return UpsertAddWithEviction, kv.value
		} else {
			l.addKey(key, value)
			return UpsertAddNoEviction, *new(V)
		}
	} else { // present in cache, just update values
		val := l.peekWithIndex(curr).value

		l.updateWithIndex(curr, value)
		l.makeRecent(curr)
		return UpsertReplace, val
	}
}

// removeOldest evicts the oldest item in the cache.
func (l *lruBase[K, V]) removeOldest() node[K, V] {
	l.addEvictions()
	return l.deleteWithIndex(l.links[0].next)
}

// removeWithIndex detaches the given element from the doubly-linked list.
func (l *lruBase[K, V]) removeWithIndex(idx int32) {
	curr := l.links[idx]

	l.links[curr.prev].next = curr.next
	l.links[curr.next].prev = curr.prev
}

// retrieveIndexWithKey retrieves the index from the hash map.
func (l *lruBase[K, V]) retrieveIndexWithKey(key K) (int32, bool) {
	curr, ok := l.hash[key]
	return curr, ok
}

// updateWithIndex updates the value of the provided index.
func (l *lruBase[K, V]) updateWithIndex(idx int32, value V) {
	l.nodes[idx].value = value
}

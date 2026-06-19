// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lrucore

import (
	"math"
	"sync"

	"github.com/justpranavrs/tlru/internal/errs"
)

// LRUCore is basic implementation of 'Least Recently Used' Cache.
// It uses a contiguous array of nodes as opposed to the standard doubly-linked list.
// It uses a hash map to track the key to the index in the array.
// It has thread-safe operations.
//
// It is recommended for key K to be primitives ([int], [uint64], [string]).
type LRUCore[K comparable, V any] struct {
	// mu is a mutual extension lock.
	mu sync.Mutex

	// hash maps the key to a unique index in the nodes array.
	hash map[K]int32

	// nodes are the elements of the doubly-linked list.
	// It is an array of [node].
	nodes []node[K, V]

	// capacity represents the maximum allocated space for the cache.
	capacity int

	// tail represents the last index in the nodes array.
	// It has a value of capacity + 1.
	tail int32

	// free points to the next available space for the cache.
	free int32
}

// node represents a single element in [LRUCore] instance's internal
// doubly-linked list. It is packed into a contiguous array
// to maximize CPU cache and avoid memory fragmentation.
type node[K comparable, V any] struct {
	// key is the identifier used to lookup in the [LRUCore] cache.
	// It is recommended for key to be primitives (int, uint64, string).
	key K

	// value holds the actual data stored in the cache
	value V

	// prev holds the index to the previous element in the doubly-linked list.
	prev int32

	// next holds the index to the next element in the doubly-linked list.
	next int32
}

// New creates an instance of [LRUCore] using the given capacity.
//
// Returns an [errs.ErrCoreInvalidCapacity] if the capacity is not in [2, 2147483646].
func New[K comparable, V any](capacity int) (*LRUCore[K, V], error) {
	if capacity < 2 || (capacity > math.MaxInt32-2) { // For capacity 1, a variable can be used.
		return nil, errs.ErrCoreInvalidCapacity
	}
	tail := 1 + int32(capacity)

	// allocate the initial nodes array with a size of 2+cap
	nodes := make([]node[K, V], 2+capacity)
	nodes[0].next = tail
	nodes[tail].prev = 0

	// lru
	lru := &LRUCore[K, V]{
		hash:     make(map[K]int32, capacity),
		nodes:    nodes,
		capacity: capacity,
		tail:     tail,
		free:     0,
	}
	return lru, nil
}

// Capacity returns the maximum allocated capacity of the LRU cache.
func (l *LRUCore[K, V]) Capacity() int {
	return l.capacity
}

// Compaction is done to avoid memory fragmentation in the nodes array over time.
// It is done by copying all the elements in the array in order to hit
// upcoming operations in L1/L2 cache lines.
//
// This is an experimental feature and has to be tested if it will
// produce better results.
func (l *LRUCore[K, V]) Compaction() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.hash) <= 1 { // no compaction can be done
		return
	}
	out := make([]node[K, V], 2+l.capacity)

	curr := l.nodes[0].next
	idx := int32(1)
	for curr != l.tail { // the compaction loop
		next := l.nodes[curr].next
		out[idx] = l.nodes[curr]

		out[idx].prev = idx - 1
		out[idx].next = idx + 1

		l.hash[out[idx].key] = idx
		curr = next
		idx++
	}
	out[0] = l.nodes[0] // head
	out[0].next = 1

	out[l.tail] = l.nodes[l.tail] // tail
	out[l.tail].prev = idx - 1
	out[idx-1].next = l.tail

	if len(l.hash) < l.capacity { // set free index
		l.free = idx - 1
	} else {
		l.free = -1
	}

	// reset disorder and set index 1 as lru
	l.nodes = out
}

// Contains verifies if the key is present in the LRU Cache.
// It does not update the key to recent status.
func (l *LRUCore[K, V]) Contains(key K) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.hash[key]
	return ok
}

// Flush clears the LRU cache of all its keys and values.
func (l *LRUCore[K, V]) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for k := range l.hash {
		delete(l.hash, k)
	}

	l.nodes[0].next = l.tail
	l.nodes[l.tail].prev = 0
	l.free = 0
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
func (l *LRUCore[K, V]) Get(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	curr, ok := l.hash[key]
	if !ok {
		return *new(V), false // not present in cache
	}
	if curr != l.nodes[l.tail].prev { // not already recent
		l.remove(curr)
		l.makeRecent(curr)
	}
	val := l.nodes[curr].value
	return val, true
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found.
func (l *LRUCore[K, V]) Peek(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	curr, ok := l.hash[key]
	if !ok {
		return *new(V), false
	}
	return l.nodes[curr].value, true
}

// Put adds a new value to the cache with the given key.
func (l *LRUCore[K, V]) Put(key K, value V) {
	l.PutGrew(key, value)
}

// PutGrew adds a new value to the cache with the given key.
// It returns true if size of the cache increased, else returns false.
func (l *LRUCore[K, V]) PutGrew(key K, value V) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	var grew bool
	curr, ok := l.hash[key]
	if !ok { // not present in cache
		if len(l.hash) == l.capacity {
			l.removeOld()
			grew = false
		} else {
			l.free++
			grew = true
		}
		l.addKey(key, value)
	} else { // present in cache, just update values
		l.nodes[curr].value = value
		if curr != l.nodes[l.tail].prev { // not already recent
			l.remove(curr)
			l.makeRecent(curr)
		}
		grew = false
	}
	return grew
}

// Size returns the current size of the LRU cache.
// It is a lock-free operation because it directly
// fetches from the memory.
func (l *LRUCore[K, V]) Size() int {
	return len(l.hash)
}

// addKey adds the cache with the given key and cache value.
func (l *LRUCore[K, V]) addKey(key K, value V) {
	l.nodes[l.free] = node[K, V]{
		key:   key,
		value: value,
		prev:  l.nodes[l.tail].prev,
		next:  l.tail,
	}
	l.nodes[l.nodes[l.tail].prev].next = l.free
	l.nodes[l.tail].prev = l.free

	l.hash[key] = l.free
}

// makeRecent sets the key as 'Most Recently Used'.
func (l *LRUCore[K, V]) makeRecent(idx int32) {
	l.nodes[l.nodes[l.tail].prev].next = idx
	l.nodes[idx].prev = l.nodes[l.tail].prev

	l.nodes[idx].next = l.tail
	l.nodes[l.tail].prev = idx
}

// remove detaches the given element from the doubly-linked list.
func (l *LRUCore[K, V]) remove(idx int32) {
	curr := l.nodes[idx]

	l.nodes[curr.prev].next = curr.next
	l.nodes[curr.next].prev = curr.prev
}

// removeOld removes the 'Least Recently Used' cache.
func (l *LRUCore[K, V]) removeOld() {
	old := l.nodes[0].next
	l.remove(old)

	delete(l.hash, l.nodes[old].key)
	l.free = old // set the next free index to the dropped cache value
}

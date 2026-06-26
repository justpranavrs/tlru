// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lrucore

import "sync"

// syncBase is the mutex-locked implementation of [base].
// It has thread-safe operations.
type syncBase[K comparable, V any, C Shard[K, V]] struct {
	// mu is a mutual extension lock.
	mu sync.Mutex

	lru C
}

// Capacity returns the maximum allocated capacity of the LRU cache.
func (l *syncBase[K, V, C]) Capacity() int {
	return l.lru.Capacity()
}

// Contains checks whether the key is present in the Cache.
func (l *syncBase[K, V, C]) Contains(key K) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Contains(key)
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *syncBase[K, V, C]) Delete(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Delete(key)
}

// Flush clears the LRU cache of all its keys and values.
func (l *syncBase[K, V, C]) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lru.Flush()
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
func (l *syncBase[K, V, C]) Get(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Get(key)
}

// GetMany allows retrieval of multiple keys at the same time.
//
// The keys, values and exists array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
//
// The operation modifies values and exists in-place. If a key is not present in the cache,
// the corresponding index in exists is set to false and leaves the value at that index unchanged.
func (l *syncBase[K, V, C]) GetMany(keys []K, values []V, exists []bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.GetMany(keys, values, exists)
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found.
func (l *syncBase[K, V, C]) Peek(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Peek(key)
}

// Put adds a new value to the cache with the given key.
// See [LRU.Upsert] for detailed information on cache state transitions.
func (l *syncBase[K, V, C]) Put(key K, value V) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lru.Put(key, value)
}

// PutMany allows the addition of multiple key-value pairs at the
// same time.
//
// The keys and values array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
func (l *syncBase[K, V, C]) PutMany(keys []K, values []V) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.PutMany(keys, values)
}

// ResetStats resets the stats of the LRU cache.
func (l *syncBase[K, V, C]) ResetStats() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lru.ResetStats()
}

// Size returns the current size of the LRU cache.
func (l *syncBase[K, V, C]) Size() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Size()
}

// Stats return the current stats of the LRU cache.
func (l *syncBase[K, V, C]) Stats() Stats {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Stats()
}

// Upsert adds a new value to the cache with the given key.
// It returns [UpsertState] based on how the internal state of the cache changed.
//
// It also returns a value based on [UpsertState]
//   - [AddNoEvict] returns the zero value of V.
//   - [AddOnEvict] returns the evicted value.
//   - [Replace] returns the old value the key had.
func (l *syncBase[K, V, C]) Upsert(key K, value V) (UpsertState, V) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Upsert(key, value)
}

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"sync"
	"time"
)

// syncBase is the mutex-locked implementation of [Shard].
// It has thread-safe operations.
type syncBase[K comparable, V any, S Shard[K, V]] struct {
	// mu is a mutual extension lock.
	mu sync.Mutex

	// the main base lockless lru cache.
	lru S
}

// Capacity returns the maximum allocated capacity of the LRU cache.
func (l *syncBase[K, V, S]) Capacity() int {
	return l.lru.Capacity()
}

// Close safely terminates the cache instance and frees up the memory.
func (l *syncBase[K, V, S]) Close() {
	l.lru.Close()
}

// Contains checks whether the key is present in the Cache.
func (l *syncBase[K, V, S]) Contains(key K) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Contains(key)
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *syncBase[K, V, S]) Delete(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Delete(key)
}

// Flush clears the LRU cache of all its keys and values.
func (l *syncBase[K, V, S]) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lru.Flush()
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
func (l *syncBase[K, V, S]) Get(key K) (V, bool) {
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
func (l *syncBase[K, V, S]) GetMany(keys []K, values []V, exists []bool) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.GetMany(keys, values, exists)
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found.
func (l *syncBase[K, V, S]) Peek(key K) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Peek(key)
}

// Put adds a new value to the cache with the given key.
// See [LRU.Upsert] for detailed information on cache state transitions.
func (l *syncBase[K, V, S]) Put(key K, value V) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lru.Put(key, value)
}

// PutMany allows the addition of multiple key-value pairs at the
// same time.
//
// The keys and values array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
func (l *syncBase[K, V, S]) PutMany(keys []K, values []V) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.PutMany(keys, values)
}

// ResetStats resets the stats of the LRU cache.
func (l *syncBase[K, V, S]) ResetStats() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lru.ResetStats()
}

// Size returns the current size of the LRU cache.
func (l *syncBase[K, V, S]) Size() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Size()
}

// Stats return the current stats of the LRU cache.
func (l *syncBase[K, V, S]) Stats() Stats {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Stats()
}

// Upsert adds a new value to the cache with the given key.
// It returns [UpsertState] based on how the internal state of the cache changed.
//
// It also returns a value based on [UpsertState]
//   - [UpsertAddNoEviction] returns the zero value of V.
//   - [UpsertAddWithEviction] returns the evicted value.
//   - [UpsertReplace] returns the old value the key had.
func (l *syncBase[K, V, S]) Upsert(key K, value V) (UpsertState, V) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.lru.Upsert(key, value)
}

// tSyncBase is the mutex-locked implementation of [TTLShard].
// It has thread-safe operations.
//
// tSyncBase does not implement syncBase because GoDoc doesn't recognize multiple
// levels of struct embedding even though compiler can.
type tSyncBase[K comparable, V any, T TTLShard[K, V]] struct {
	// mu is a mutual extension lock.
	mu sync.Mutex

	// the main base lockless lru cache.
	lru T
}

// Capacity returns the maximum allocated capacity of the LRU cache.
func (t *tSyncBase[K, V, S]) Capacity() int {
	return t.lru.Capacity()
}

// Close safely closes the background clock when TTL is enabled on the cache.
func (t *tSyncBase[K, V, S]) Close() {
	t.lru.Close()
}

// Contains checks whether the key is present in the Cache.
func (t *tSyncBase[K, V, S]) Contains(key K) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.Contains(key)
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (t *tSyncBase[K, V, S]) Delete(key K) (V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.Delete(key)
}

// Flush clears the LRU cache of all its keys and values.
func (t *tSyncBase[K, V, S]) Flush() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lru.Flush()
}

// Get retrieves the cache value using key.
// It returns false if the key is not found or has been expired.
func (t *tSyncBase[K, V, S]) Get(key K) (V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.Get(key)
}

// GetMany allows retrieval of multiple keys at the same time.
//
// The keys, values and exists array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
//
// The operation modifies values and exists in-place. If a key is not present in the cache or
// has been expired, the corresponding index in exists is set to false
// and leaves the value at that index unchanged.
func (t *tSyncBase[K, V, S]) GetMany(keys []K, values []V, exists []bool) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.GetMany(keys, values, exists)
}

// Check [TLRU.Get] on how Get works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (t *tSyncBase[K, V, S]) GetWithTTL(key K) (V, time.Duration, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.GetWithTTL(key)
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found or if it has been expired.
func (t *tSyncBase[K, V, S]) Peek(key K) (V, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.Peek(key)
}

// Check [TLRU.Peek] on how Peek works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (t *tSyncBase[K, V, S]) PeekWithTTL(key K) (V, time.Duration, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.PeekWithTTL(key)
}

// Put adds a new value to the cache with the given key and assigns a new
// timestamp to the key using the default TTL.
//
// See [TLRU.Upsert] for detailed information on cache state transitions.
//
// See [TLRU.PutWithTTL] for adding keys with custom TTL.
func (t *tSyncBase[K, V, S]) Put(key K, value V) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lru.Put(key, value)
}

// PutMany allows the addition of multiple key-value pairs and also assigns new timestamps
// for the keys, at the same time.
//
// The keys and values array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
func (t *tSyncBase[K, V, S]) PutMany(keys []K, values []V) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.PutMany(keys, values)
}

// PutWithTTL adds a new value to the cache with the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (t *tSyncBase[K, V, S]) PutWithTTL(key K, value V, ttl time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lru.PutWithTTL(key, value, ttl)
}

// Refresh resets the TTL of an existing key using the default ttl.
// It returns false if the key could not be found.
func (t *tSyncBase[K, V, S]) Refresh(key K) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.Refresh(key)
}

// ResetStats resets the stats of the LRU cache.
func (t *tSyncBase[K, V, S]) ResetStats() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.lru.ResetStats()
}

// SetTTL resets the TTL of an existing key using the given ttl in the argument.
// It returns false if the key could not be found.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (t *tSyncBase[K, V, S]) SetTTL(key K, ttl time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.SetTTL(key, ttl)
}

// Size returns the current size of the LRU cache.
func (t *tSyncBase[K, V, S]) Size() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.Size()
}

// Stats return the current stats of the LRU cache.
func (t *tSyncBase[K, V, S]) Stats() Stats {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.Stats()
}

// TTL returns the remaining TTL for the key.
func (t *tSyncBase[K, V, S]) TTL(key K) (time.Duration, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.TTL(key)
}

// Upsert adds a new value to the cache with the given key and
// also updating their timestamps with the default TTL.
// It returns [UpsertState] based on how the internal state of the cache changed.
//
// It also returns a value based on [UpsertState]
//   - [UpsertAddNoEviction] returns the zero value of V.
//   - [UpsertAddWithEviction] returns the evicted value.
//   - [UpsertReplace] returns the old value the key had.
//   - [UpsertAddAfterExpiration] returns the expired value that was overwritten.
func (t *tSyncBase[K, V, S]) Upsert(key K, value V) (UpsertState, V) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.Upsert(key, value)
}

// UpsertWithTTL adds a new value to the cache with the given key and the provided ttl value.
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [TLRU.Upsert] on how Upsert works.
func (t *tSyncBase[K, V, S]) UpsertWithTTL(key K, value V, ttl time.Duration) (UpsertState, V) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.lru.UpsertWithTTL(key, value, ttl)
}

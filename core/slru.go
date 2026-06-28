// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

// slruBase is the basic implementation of Segmented LRU.
// It uses two [lruBase], a warm and a cold region to avoid cache pollution.
type slruBase[K comparable, V any] struct {
	// probationary is the cold region of the slru cache.
	probationary *lruBase[K, V]

	// protected is the warm region of slru cache.
	protected *lruBase[K, V]

	// stats represents the metrics added by its private methods
	stats Stats
}

// assembleSLRU creates an instance of [slruBase] using the given capacity and ratio.
func assembleSLRU[K comparable, V any](capacity int, ratio int) (*slruBase[K, V], error) {
	if ratio < 0 || ratio > 100 {
		return nil, ErrInvalidSLRURatio
	}

	probCap := capacity * ratio / 100
	prob, err := assembleLRU[K, V](probCap)
	if err != nil {
		return nil, err
	}

	prot, err := assembleLRU[K, V](capacity - probCap)
	if err != nil {
		return nil, err
	}

	return &slruBase[K, V]{
		probationary: prob,
		protected:    prot,
	}, nil
}

// Capacity returns the maximum allocated capacity of the SLRU cache.
func (s *slruBase[K, V]) Capacity() int {
	return s.probationary.capacity + s.protected.capacity
}

// Contains checks whether the key is present in the Cache.
func (s *slruBase[K, V]) Contains(key K) bool {
	_, ok := s.protected.peekWithKey(key)
	if !ok {
		_, ok := s.probationary.peekWithKey(key)
		return ok
	}
	return true
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (s *slruBase[K, V]) Delete(key K) (V, bool) {
	val, ok := s.protected.deleteWithKey(key)
	if !ok {
		return s.probationary.deleteWithKey(key)
	}
	return val, true
}

// Flush clears the SLRU cache of all its keys and values.
func (s *slruBase[K, V]) Flush() {
	s.protected.clearState()
	s.probationary.clearState()
}

// Get retrieves the cache value using key.
// It returns false if the key is not found.
func (s *slruBase[K, V]) Get(key K) (V, bool) {
	return s.getWithKey(key)
}

// GetMany allows retrieval of multiple keys at the same time.
//
// The keys, values and exists array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
//
// The operation modifies values and exists in-place. If a key is not present in the cache,
// the corresponding index in exists is set to false and leaves the value at that index unchanged.
func (s *slruBase[K, V]) GetMany(keys []K, values []V, exists []bool) error {
	if (len(keys) != len(values)) || (len(keys) != len(exists)) {
		return ErrInvalidBatchSize
	}

	for i := range keys {
		val, ok := s.getWithKey(keys[i])
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
func (s *slruBase[K, V]) Peek(key K) (V, bool) {
	val, ok := s.protected.peekWithKey(key)
	if !ok {
		return s.probationary.peekWithKey(key)
	}
	return val, true
}

// Put adds a new value to the cache with the given key.
// See [SLRU.Upsert] for detailed information on cache state transitions.
func (s *slruBase[K, V]) Put(key K, value V) {
	s.putWithKey(key, value)
}

// PutMany allows the addition of multiple key-value pairs at the
// same time.
//
// The keys and values array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
func (s *slruBase[K, V]) PutMany(keys []K, values []V) error {
	if len(keys) != len(values) {
		return ErrInvalidBatchSize
	}

	for i := range keys {
		s.putWithKey(keys[i], values[i])
	}
	return nil
}

// ResetStats resets the stats of the LRU cache.
func (s *slruBase[K, V]) ResetStats() {
	s.stats = Stats{}
	s.probationary.stats = Stats{}
	s.protected.stats = Stats{}
}

// Size returns the current size of the LRU cache.
func (s *slruBase[K, V]) Size() int {
	return len(s.probationary.hash) + len(s.protected.hash)
}

// Stats return the current stats of the LRU cache.
func (s *slruBase[K, V]) Stats() Stats {
	return Stats{
		Hits:        s.stats.Hits + s.probationary.stats.Hits + s.protected.stats.Hits,
		Misses:      s.stats.Misses + s.probationary.stats.Misses,
		Evictions:   s.stats.Evictions + s.probationary.stats.Evictions,
		Expirations: s.stats.Expirations + s.probationary.stats.Expirations + s.protected.stats.Expirations,
	}
}

// Upsert adds a new value to the cache with the given key.
// It returns [UpsertState] based on how the internal state of the cache changed.
//
// It also returns a value based on [UpsertState]
//   - [UpsertAddNoEviction] returns the zero value of V.
//   - [UpsertAddWithEviction] returns the evicted value.
//   - [UpsertReplace] returns the old value the key had.
func (s *slruBase[K, V]) Upsert(key K, value V) (UpsertState, V) {
	return s.putWithKey(key, value)
}

// addEvictions increases the evictions stat by 1.
func (s *slruBase[K, V]) addEvictions() {
	s.stats.Evictions++
}

// addExpirations increases the expirations stat by 1.
func (s *slruBase[K, V]) addExpirations() {
	s.stats.Expirations++
}

// addHits increases the hits stat by 1.
func (s *slruBase[K, V]) addHits() {
	s.stats.Hits++
}

// addKey adds the cache with the given key and cache value.
func (s *slruBase[K, V]) addKey(key K, value V) {
	s.probationary.addKey(key, value)
}

// addMisses increases the misses stat by 1.
func (s *slruBase[K, V]) addMisses() {
	s.stats.Misses++
}

// applyOffsetIndex, offsets the index for protected.
func (s *slruBase[K, V]) applyOffsetIndex(idx int32) int32 {
	return int32(s.probationary.capacity) + idx
}

// clearState erases the entire cache's internal state
func (s *slruBase[K, V]) clearState() {
	s.probationary.clearState()
	s.protected.clearState()
}

// deleteWithIndex removes a node from the cache.
func (s *slruBase[K, V]) deleteWithIndex(idx int32) node[K, V] {
	if idx < int32(s.probationary.capacity) {
		return s.probationary.deleteWithIndex(idx)
	} else {
		return s.protected.deleteWithIndex(s.removeOffsetIndex(idx))
	}
}

// deleteWithKey removes a key from the cache.
func (s *slruBase[K, V]) deleteWithKey(key K) (V, bool) {
	curr, ok := s.protected.retrieveIndexWithKey(key)
	if !ok {
		curr, ok := s.probationary.retrieveIndexWithKey(key)
		if !ok {
			return *new(V), false
		}
		return s.probationary.deleteWithIndex(curr).value, true
	}
	return s.protected.deleteWithIndex(curr).value, true
}

// getWithIndex updates the recency, the stats and returns the value.
func (s *slruBase[K, V]) getWithIndex(idx int32) V {
	if idx < int32(s.probationary.capacity) {
		return s.probationary.getWithIndex(idx)
	} else {
		return s.protected.getWithIndex(s.removeOffsetIndex(idx))
	}
}

// getWithKey retrieves the value of the key passed as argument.
// It returns false if the key is not present.
func (s *slruBase[K, V]) getWithKey(key K) (V, bool) {
	val, ok := s.protected.getWithKey(key)
	if !ok {
		curr, exists := s.probationary.retrieveIndexWithKey(key)
		if !exists {
			s.probationary.addMisses()
			return *new(V), false
		}
		s.probationary.addHits()

		val = s.probationary.deleteWithIndex(curr).value // this won't trigger an eviction
		s.promoteKey(key, val)
	}
	return val, true
}

// makeRecent sets the key as 'Most Recently Used'.
func (s *slruBase[K, V]) makeRecent(idx int32) {
	if idx < int32(s.probationary.capacity) {
		s.probationary.makeRecent(idx)
	} else {
		s.protected.makeRecent(s.removeOffsetIndex(idx))
	}
}

// peekWithIndex retrieves the node without updating recency.
func (s *slruBase[K, V]) peekWithIndex(idx int32) node[K, V] {
	if idx < int32(s.probationary.capacity) {
		return s.probationary.peekWithIndex(idx)
	} else {
		return s.protected.peekWithIndex(s.removeOffsetIndex(idx))
	}
}

// peekWithKey retrieves value of key without updating cache internal state.
// Returns false when key not found.
func (s *slruBase[K, V]) peekWithKey(key K) (V, bool) {
	curr, ok := s.protected.retrieveIndexWithKey(key)
	if !ok {
		curr, ok := s.probationary.retrieveIndexWithKey(key)
		if !ok {
			return *new(V), false
		}
		return s.probationary.peekWithIndex(curr).value, true
	}
	return s.protected.peekWithIndex(curr).value, true
}

// promoteKey promotes a key from probationary to protected and demotes
// a key if necessary. It ensures that the probationary key is already deleted
// to have space for the protected key.
func (s *slruBase[K, V]) promoteKey(key K, value V) {
	if len(s.protected.hash) == s.protected.capacity { // demoted key
		kv := s.protected.removeOldest() // this will trigger an eviction in protected
		s.probationary.addKey(kv.key, kv.value)
	}
	s.protected.addKey(key, value) // promoted key
}

// putWithKey adds or updates the cache value with key "key".
// It also returns a value based on whether it was evicted or replaced.
func (s *slruBase[K, V]) putWithKey(key K, value V) (UpsertState, V) {
	curr, ok := s.protected.retrieveIndexWithKey(key)
	if !ok {
		curr, exists := s.probationary.retrieveIndexWithKey(key)
		if !exists {
			return s.probationary.putNewKey(key, value)
		}
		val := s.probationary.deleteWithIndex(curr).value // this won't trigger an eviction
		s.promoteKey(key, value)
		return UpsertReplace, val
	}
	return s.protected.putOldKey(curr, value)
}

// removeOffsetIndex, offsets the index for protected.
func (s *slruBase[K, V]) removeOffsetIndex(idx int32) int32 {
	return idx - int32(s.probationary.capacity)
}

// removeOldest evicts the oldest item in the cache.
func (s *slruBase[K, V]) removeOldest() node[K, V] {
	return s.probationary.removeOldest()
}

// removeWithIndex detaches the given element from the doubly-linked list.
func (s *slruBase[K, V]) removeWithIndex(idx int32) {
	if idx < int32(s.probationary.capacity) {
		s.probationary.removeWithIndex(idx)
	} else {
		s.protected.removeWithIndex(s.removeOffsetIndex(idx))
	}
}

// retrieveIndexWithKey retrieves the index from the hash map.
func (s *slruBase[K, V]) retrieveIndexWithKey(key K) (int32, bool) {
	curr, ok := s.protected.retrieveIndexWithKey(key)
	if !ok {
		return s.probationary.retrieveIndexWithKey(key)
	}
	return s.applyOffsetIndex(curr), true
}

// updateWithIndex updates the value of the provided index.
func (s *slruBase[K, V]) updateWithIndex(idx int32, value V) {
	if idx < int32(s.probationary.capacity) {
		s.probationary.updateWithIndex(idx, value)
	} else {
		s.protected.updateWithIndex(s.removeOffsetIndex(idx), value)
	}
}

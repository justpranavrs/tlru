// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

type slruBase[K comparable, V any] struct {
	// probationary is the cold region of the slru cache.
	probationary *lruBase[K, V]

	// protected is the warm region of slru cache.
	protected *lruBase[K, V]

	// stats represents the metrics added by its private methods
	stats Stats
}

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

func (s *slruBase[K, V]) Capacity() int {
	return s.probationary.capacity + s.protected.capacity
}

func (s *slruBase[K, V]) Contains(key K) bool {
	_, ok := s.protected.peekWithKey(key)
	if !ok {
		_, ok := s.protected.peekWithKey(key)
		return ok
	}
	return true
}

func (s *slruBase[K, V]) Delete(key K) (V, bool) {
	val, ok := s.protected.deleteWithKey(key)
	if !ok {
		return s.probationary.deleteWithKey(key)
	}
	return val, true
}

func (s *slruBase[K, V]) Flush() {
	s.protected.clearState()
	s.probationary.clearState()
}

func (s *slruBase[K, V]) Get(key K) (V, bool) {
	return s.getWithKey(key)
}

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

func (s *slruBase[K, V]) Peek(key K) (V, bool) {
	val, ok := s.protected.peekWithKey(key)
	if !ok {
		return s.probationary.peekWithKey(key)
	}
	return val, true
}

func (s *slruBase[K, V]) Put(key K, value V) {
	s.putWithKey(key, value)
}

func (s *slruBase[K, V]) PutMany(keys []K, values []V) error {
	if len(keys) != len(values) {
		return ErrInvalidBatchSize
	}

	for i := range keys {
		s.putWithKey(keys[i], values[i])
	}
	return nil
}

func (s *slruBase[K, V]) ResetStats() {
	s.stats = Stats{}
	s.probationary.stats = Stats{}
	s.protected.stats = Stats{}
}

func (s *slruBase[K, V]) Size() int {
	return len(s.probationary.hash) + len(s.protected.hash)
}

func (s *slruBase[K, V]) Stats() Stats {
	return Stats{
		Hits:        s.stats.Hits + s.probationary.stats.Hits + s.protected.stats.Hits,
		Misses:      s.stats.Misses + s.probationary.stats.Misses,
		Evictions:   s.stats.Evictions + s.probationary.stats.Evictions,
		Expirations: s.stats.Expirations + s.probationary.stats.Expirations + s.protected.stats.Expirations,
	}
}

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
		return s.protected.deleteWithIndex(s.offsetIndex(idx))
	}
}

// deleteWithKey removes a key from the cache.
func (s *slruBase[K, V]) deleteWithKey(key K) (V, bool) {
	curr, ok := s.protected.retrieveIndexWithKey(key)
	if !ok {
		curr, ok := s.protected.retrieveIndexWithKey(key)
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
		return s.protected.getWithIndex(s.offsetIndex(idx))
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

		val = s.probationary.deleteWithIndex(curr).value   // this won't trigger an eviction
		if len(s.protected.hash) == s.protected.capacity { // demoted key
			kv := s.protected.removeOldest() // this will trigger an eviction in protected
			s.probationary.addKey(kv.key, kv.value)
		}
		s.protected.addKey(key, val) // promoted key
	}
	return val, true
}

// makeRecent sets the key as 'Most Recently Used'.
func (s *slruBase[K, V]) makeRecent(idx int32) {
	if idx < int32(s.probationary.capacity) {
		s.probationary.makeRecent(idx)
	} else {
		s.protected.makeRecent(s.offsetIndex(idx))
	}
}

// peekWithIndex retrieves the node without updating recency.
func (s *slruBase[K, V]) peekWithIndex(idx int32) node[K, V] {
	if idx < int32(s.probationary.capacity) {
		return s.probationary.peekWithIndex(idx)
	} else {
		return s.protected.peekWithIndex(s.offsetIndex(idx))
	}
}

// peekWithKey retrieves value of key without updating cache internal state.
// Returns false when key not found.
func (s *slruBase[K, V]) peekWithKey(key K) (V, bool) {
	curr, ok := s.protected.retrieveIndexWithKey(key)
	if !ok {
		curr, ok := s.protected.retrieveIndexWithKey(key)
		if !ok {
			return *new(V), false
		}
		return s.probationary.peekWithIndex(curr).value, true
	}
	return s.protected.peekWithIndex(curr).value, true
}

// offsetIndex, offsets the index for protected.
func (s *slruBase[K, V]) offsetIndex(idx int32) int32 {
	return int32(s.probationary.capacity) + idx
}

// putWithKey adds or updates the cache value with key "key".
// It also returns a value based on whether it was evicted or replaced.
func (s *slruBase[K, V]) putWithKey(key K, value V) (UpsertState, V) {
	curr, ok := s.protected.retrieveIndexWithKey(key)
	if !ok {
		return s.probationary.putWithKey(key, value)
	}
	val := s.protected.peekWithIndex(curr).value

	s.protected.updateWithIndex(curr, value)
	s.protected.makeRecent(curr)
	return Replace, val
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
		s.protected.removeWithIndex(s.offsetIndex(idx))
	}
}

// retrieveIndexWithKey retrieves the index from the hash map.
func (s *slruBase[K, V]) retrieveIndexWithKey(key K) (int32, bool) {
	curr, ok := s.protected.retrieveIndexWithKey(key)
	if !ok {
		return s.probationary.retrieveIndexWithKey(key)
	}
	return s.offsetIndex(curr), true
}

// updateWithIndex updates the value of the provided index.
func (s *slruBase[K, V]) updateWithIndex(idx int32, value V) {
	if idx < int32(s.probationary.capacity) {
		s.probationary.updateWithIndex(idx, value)
	} else {
		s.protected.updateWithIndex(s.offsetIndex(idx), value)
	}
}

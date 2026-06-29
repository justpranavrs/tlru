// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"time"

	"github.com/justpranavrs/tlru/clock"
)

// tlruBase extends [lruBase] for TTL support. It is not safe
// for concurrent workloads. [syncBase] provides the [sync.Mutex]
// locks for it.
type tlruBase[K comparable, V any, B base[K, ttlValue[V]]] struct {
	// base represents the main cache which holds the keys and values.
	base B

	// clock is the background timer to ensure fast time loads
	// without halting the LRU operations.
	clock *clock.Clock

	// ttl determines the default (time-to-live) duration of an element.
	ttl int64

	// sliding represents if Sliding TTL is enabled.
	// If false Absolute TTL is enabled.
	sliding bool
}

// ttlValue is the wrapper for [LRU] with a timestamp and the value.
type ttlValue[V any] struct {
	// expiresAt is the time when the key would get expired.
	expiresAt int64

	// value is the actual data of the cache.
	value V
}

// makeTLRU creates an instance of [tlruBase] using the given base and sets
// the default expiration timer based on the argument "ttl".
func makeTLRU[K comparable, V any, B base[K, ttlValue[V]]](base B, ttl time.Duration, clock *clock.Clock, sliding bool) *tlruBase[K, V, B] {
	return &tlruBase[K, V, B]{
		base:    base,
		clock:   clock,
		ttl:     clock.Ticks(ttl),
		sliding: sliding,
	}
}

// Capacity returns the maximum allocated capacity of the LRU cache.
func (l *tlruBase[K, V, B]) Capacity() int {
	return l.base.Capacity()
}

// Close safely closes the background clock when TTL is enabled and also frees up
// memory on the cache.
func (l *tlruBase[K, V, B]) Close() {
	if l.clock != nil {
		l.clock.Stop()
	}
	l.base.Close()
}

// Contains checks whether the key is present in the Cache.
func (l *tlruBase[K, V, B]) Contains(key K) bool {
	_, _, ok := l.peekWithKey(key)
	return ok
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *tlruBase[K, V, B]) Delete(key K) (V, bool) {
	val, ok := l.base.deleteWithKey(key)
	return val.value, ok
}

// Flush clears the LRU cache of all its keys and values.
func (l *tlruBase[K, V, B]) Flush() {
	l.base.Flush()
}

// Get retrieves the cache value using key.
// It returns false if the key is not found or has been expired.
func (l *tlruBase[K, V, B]) Get(key K) (V, bool) {
	val, _, ok := l.getWithKey(key)
	return val, ok
}

// GetMany allows retrieval of multiple keys at the same time.
//
// The keys, values and exists array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
//
// The operation modifies values and exists in-place. If a key is not present in the cache or
// has been expired, the corresponding index in exists is set to false
// and leaves the value at that index unchanged.
func (l *tlruBase[K, V, B]) GetMany(keys []K, values []V, exists []bool) error {
	if (len(keys) != len(values)) || (len(keys) != len(exists)) {
		return ErrInvalidBatchSize
	}

	for i := range keys {
		val, _, ok := l.getWithKey(keys[i])
		exists[i] = ok
		if ok {
			values[i] = val
		}
	}
	return nil
}

// Check [TLRU.Get] on how Get works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *tlruBase[K, V, B]) GetWithTTL(key K) (V, time.Duration, bool) {
	return l.getWithKey(key)
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found or if it has been expired.
func (l *tlruBase[K, V, B]) Peek(key K) (V, bool) {
	val, _, ok := l.peekWithKey(key)
	return val, ok
}

// Check [TLRU.Peek] on how Peek works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *tlruBase[K, V, B]) PeekWithTTL(key K) (V, time.Duration, bool) {
	return l.peekWithKey(key)
}

// Put adds a new value to the cache with the given key and assigns a new
// timestamp to the key using the default TTL.
//
// See [TLRU.Upsert] for detailed information on cache state transitions.
//
// See [TLRU.PutWithTTL] for adding keys with custom TTL.
func (l *tlruBase[K, V, B]) Put(key K, value V) {
	l.putWithKey(key, value, l.ttl)
}

// PutMany allows the addition of multiple key-value pairs and also assigns new timestamps
// for the keys, at the same time.
//
// The keys and values array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
func (l *tlruBase[K, V, B]) PutMany(keys []K, values []V) error {
	if len(keys) != len(values) {
		return ErrInvalidBatchSize
	}

	for i := range keys {
		l.putWithKey(keys[i], values[i], l.ttl)
	}
	return nil
}

// PutWithTTL adds a new value to the cache with the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (l *tlruBase[K, V, B]) PutWithTTL(key K, value V, ttl time.Duration) {
	l.putWithKey(key, value, l.clock.Ticks(ttl))
}

// Refresh resets the TTL of an existing key using the default ttl.
// It returns false if the key could not be found.
func (l *tlruBase[K, V, B]) Refresh(key K) bool {
	return l.refreshWithKey(key, l.ttl)
}

// ResetStats resets the stats of the LRU cache.
func (l *tlruBase[K, V, B]) ResetStats() {
	l.base.ResetStats()
}

// SetTTL resets the TTL of an existing key using the given ttl in the argument.
// It returns false if the key could not be found.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (l *tlruBase[K, V, B]) SetTTL(key K, ttl time.Duration) bool {
	return l.refreshWithKey(key, l.clock.Ticks(ttl))
}

// Size returns the current size of the LRU cache.
func (l *tlruBase[K, V, B]) Size() int {
	return l.base.Size()
}

// Stats return the current stats of the LRU cache.
func (l *tlruBase[K, V, B]) Stats() Stats {
	return l.base.Stats()
}

// TTL returns the remaining TTL for the key.
func (l *tlruBase[K, V, B]) TTL(key K) (time.Duration, bool) {
	_, ttl, ok := l.peekWithKey(key)
	return ttl, ok
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
func (l *tlruBase[K, V, B]) Upsert(key K, value V) (UpsertState, V) {
	return l.putWithKey(key, value, l.ttl)
}

// UpsertWithTTL adds a new value to the cache with the given key and the provided ttl value.
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [TLRU.Upsert] on how Upsert works.
func (l *tlruBase[K, V, B]) UpsertWithTTL(key K, value V, ttl time.Duration) (UpsertState, V) {
	return l.putWithKey(key, value, l.clock.Ticks(ttl))
}

// expireWithIndex verifies if the timestamp has expired. If it has, it will evict the key.
func (l *tlruBase[K, V, B]) expireWithIndex(idx int32, val ttlValue[V]) (V, time.Duration, bool) {
	if val.expiresAt <= l.clock.Now() {
		l.base.deleteWithIndex(idx)
		l.base.addExpirations()

		l.base.addMisses()
		return *new(V), 0, false
	}
	if l.sliding {
		l.updateTTLWithIndex(idx, l.ttl)
	}
	return val.value, l.clock.Duration() * time.Duration(l.clock.Until(val.expiresAt)), true
}

// getWithKey retrieves the value and also removes the key if the key has expired.
func (l *tlruBase[K, V, B]) getWithKey(key K) (V, time.Duration, bool) {
	curr, ok := l.base.retrieveIndexWithKey(key)
	if !ok {
		return *new(V), 0, false // not present in cache
	}
	val, ttl, ok := l.expireWithIndex(curr, l.base.peekWithIndex(curr).value)
	if ok {
		l.base.getWithIndex(curr)
	}
	return val, ttl, ok
}

// peekKey retrieves the value based on the key provided in the argument,
// without ever changing the internal state of the cache unless the key is expired.
func (l *tlruBase[K, V, B]) peekWithKey(key K) (V, time.Duration, bool) {
	curr, ok := l.base.retrieveIndexWithKey(key)
	if !ok {
		return *new(V), 0, false // not present in cache
	}
	return l.expireWithIndex(curr, l.base.peekWithIndex(curr).value)
}

// putKey inserts the new key and value into the cache. It returns how
// the internal state was updated and returns a value based on that state.
// It also adds up expiration stat, if the UpsertReplaced key was expired.
// It also takes in an argument ttl to set a custom expiration time.
func (l *tlruBase[K, V, B]) putWithKey(key K, value V, ttl int64) (UpsertState, V) {
	state, val := l.base.putWithKey(key, ttlValue[V]{
		expiresAt: l.clock.Now() + ttl,
		value:     value,
	})
	if state == UpsertReplace {
		if val.expiresAt <= l.clock.Now() {
			l.base.addExpirations()
			return UpsertAddAfterExpiration, val.value
		}
		return UpsertReplace, val.value
	}
	return state, val.value
}

// refreshWithKey updates the ttl value with the key.
func (l *tlruBase[K, V, B]) refreshWithKey(key K, ttl int64) bool {
	curr, ok := l.base.retrieveIndexWithKey(key)
	if !ok {
		return false // not present in cache
	}
	l.updateTTLWithIndex(curr, ttl)
	return true
}

// updateTTLWithIndex updates the ttl value with the index.
func (l *tlruBase[K, V, B]) updateTTLWithIndex(idx int32, ttl int64) {
	l.base.updateWithIndex(idx, ttlValue[V]{
		expiresAt: l.clock.Now() + ttl,
		value:     l.base.peekWithIndex(idx).value.value,
	})
}

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lrucore

import (
	"time"

	"github.com/justpranavrs/tlru/lruclock"
)

// TTLShard defines the blueprint of a 'Time-Aware Least Recently Used'.
// It implements [Shard].
type TTLShard[K comparable, V any] interface {
	Shard[K, V]

	// Check [Shard.Get] on how Get works.
	// It also returns the remaining TTL in the key if it was found in the cache.
	GetWithTTL(key K) (V, time.Duration, bool)

	// Check [Shard.Peek] on how Peek works.
	// It also returns the remaining TTL in the key if it was found in the cache.
	PeekWithTTL(key K) (V, time.Duration, bool)

	// PutWithTTL adds a new value to the cache with the provided ttl value.
	PutWithTTL(key K, value V, ttl time.Duration)

	// Refresh resets the TTL of an existing key using the default ttl.
	// It returns false if the key could not be found.
	Refresh(key K) bool

	// SetTTL resets the TTL of an existing key using the given ttl in the argument.
	// It returns false if the key could not be found.
	SetTTL(key K, ttl time.Duration) bool

	// TTL returns the remaining TTL for the key.
	TTL(key K) (time.Duration, bool)

	// UpsertWithTTL adds a new value to the cache with the given key and the provided ttl value.
	//
	// Check [Shard.Upsert] on how Upsert works.
	UpsertWithTTL(key K, value V, ttl time.Duration) (UpsertState, V)
}

// tlruBase extends [base] for TTL support. It is not safe
// for concurrent workloads.
type tlruBase[K comparable, V any] struct {
	// base represents the main cache which holds the keys and values.
	base *base[K, ttlValue[V]]

	// clock is the background timer to ensure fast time loads
	// without halting the LRU operations.
	clock *lruclock.Clock

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

// assembleTLRU creates an instance of [tlruBase] using the given capacity and sets
// the default expiration timer based on the argument "ttl".
func assembleTLRU[K comparable, V any](capacity int, ttl time.Duration, clock *lruclock.Clock, sliding bool) (*tlruBase[K, V], error) {
	base, err := assembleBase[K, ttlValue[V]](capacity)
	if err != nil {
		return nil, err
	}

	return &tlruBase[K, V]{
		base:    base,
		clock:   clock,
		ttl:     clock.Ticks(ttl),
		sliding: sliding,
	}, nil
}

// Capacity returns the maximum allocated capacity of the LRU cache.
func (l *tlruBase[K, V]) Capacity() int {
	return l.base.Capacity()
}

// Contains checks whether the key is present in the Cache.
func (l *tlruBase[K, V]) Contains(key K) bool {
	_, _, ok := l.peekKey(key)
	return ok
}

// Close safely closes the background clock when TTL is enabled on the cache.
func (l *tlruBase[K, V]) Close() {
	if l.clock != nil {
		l.clock.Stop()
	}
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *tlruBase[K, V]) Delete(key K) (V, bool) {
	val, ok := l.base.Delete(key)
	return val.value, ok
}

// Flush clears the LRU cache of all its keys and values.
func (l *tlruBase[K, V]) Flush() {
	l.base.Flush()
}

// Get retrieves the cache value using key.
// It returns false if the key is not found or has been expired.
func (l *tlruBase[K, V]) Get(key K) (V, bool) {
	val, _, ok := l.getKey(key)
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
func (l *tlruBase[K, V]) GetMany(keys []K, values []V, exists []bool) error {
	if (len(keys) != len(values)) || (len(keys) != len(exists)) {
		return ErrInvalidBatchSize
	}

	for i := range keys {
		val, _, ok := l.getKey(keys[i])
		exists[i] = ok
		if ok {
			values[i] = val
		}
	}
	return nil
}

// Check [TLRU.Get] on how Get works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *tlruBase[K, V]) GetWithTTL(key K) (V, time.Duration, bool) {
	return l.getKey(key)
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found or if it has been expired.
func (l *tlruBase[K, V]) Peek(key K) (V, bool) {
	val, _, ok := l.peekKey(key)
	return val, ok
}

// Check [TLRU.Peek] on how Peek works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *tlruBase[K, V]) PeekWithTTL(key K) (V, time.Duration, bool) {
	return l.peekKey(key)
}

// Put adds a new value to the cache with the given key and assigns a new
// timestamp to the key using the default TTL.
//
// See [TLRU.Upsert] for detailed information on cache state transitions.
//
// See [TLRU.PutWithTTL] for adding keys with custom TTL.
func (l *tlruBase[K, V]) Put(key K, value V) {
	l.putKey(key, value, l.ttl)
}

// PutMany allows the addition of multiple key-value pairs and also assigns new timestamps
// for the keys, at the same time.
//
// The keys and values array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
func (l *tlruBase[K, V]) PutMany(keys []K, values []V) error {
	if len(keys) != len(values) {
		return ErrInvalidBatchSize
	}

	for i := range keys {
		l.putKey(keys[i], values[i], l.ttl)
	}
	return nil
}

// PutWithTTL adds a new value to the cache with the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (l *tlruBase[K, V]) PutWithTTL(key K, value V, ttl time.Duration) {
	l.putKey(key, value, l.clock.Ticks(ttl))
}

// Refresh resets the TTL of an existing key using the default ttl.
// It returns false if the key could not be found.
func (l *tlruBase[K, V]) Refresh(key K) bool {
	return l.refreshKey(key, l.ttl)
}

// ResetStats resets the stats of the LRU cache.
func (l *tlruBase[K, V]) ResetStats() {
	l.base.ResetStats()
}

// SetTTL resets the TTL of an existing key using the given ttl in the argument.
// It returns false if the key could not be found.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (l *tlruBase[K, V]) SetTTL(key K, ttl time.Duration) bool {
	return l.refreshKey(key, l.clock.Ticks(ttl))
}

// Size returns the current size of the LRU cache.
func (l *tlruBase[K, V]) Size() int {
	return l.base.Size()
}

// Stats return the current stats of the LRU cache.
func (l *tlruBase[K, V]) Stats() Stats {
	return l.base.Stats()
}

// TTL returns the remaining TTL for the key.
func (l *tlruBase[K, V]) TTL(key K) (time.Duration, bool) {
	_, ttl, ok := l.peekKey(key)
	return ttl, ok
}

// Upsert adds a new value to the cache with the given key and
// also updating their timestamps with the default TTL.
// It returns [UpsertState] based on how the internal state of the cache changed.
//
// It also returns a value based on [UpsertState]
//   - [AddNoEvict] returns the zero value of V.
//   - [AddOnEvict] returns the evicted value.
//   - [Replace] returns the old value the key had.
//   - [AddAfterExpiration] returns the expired value that was overwritten.
func (l *tlruBase[K, V]) Upsert(key K, value V) (UpsertState, V) {
	return l.putKey(key, value, l.ttl)
}

// UpsertWithTTL adds a new value to the cache with the given key and the provided ttl value.
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [TLRU.Upsert] on how Upsert works.
func (l *tlruBase[K, V]) UpsertWithTTL(key K, value V, ttl time.Duration) (UpsertState, V) {
	return l.putKey(key, value, l.clock.Ticks(ttl))
}

// expireKey verifies if the timestamp has expired. If it has, it will evict the key.
func (l *tlruBase[K, V]) expireKey(idx int32, val ttlValue[V]) (V, time.Duration, bool) {
	if val.expiresAt <= l.clock.Now() {
		l.base.deleteKey(idx)
		l.base.stats.Expirations++

		l.base.stats.Misses++
		return *new(V), 0, false
	}
	if l.sliding {
		l.base.nodes[idx].value.expiresAt = l.clock.Now() + l.ttl
	}
	return val.value, l.clock.Duration() * time.Duration(l.clock.Until(val.expiresAt)), true
}

// getKey retrieves the value and also removes the key if the key has expired.
func (l *tlruBase[K, V]) getKey(key K) (V, time.Duration, bool) {
	curr, ok := l.base.hash[key]
	if !ok {
		return *new(V), 0, false // not present in cache
	}
	return l.expireKey(curr, l.base.getIndex(curr))
}

// peekKey retrieves the value based on the key provided in the argument,
// without ever changing the internal state of the cache unless the key is expired.
func (l *tlruBase[K, V]) peekKey(key K) (V, time.Duration, bool) {
	curr, ok := l.base.hash[key]
	if !ok {
		return *new(V), 0, false // not present in cache
	}
	return l.expireKey(curr, l.base.nodes[curr].value)
}

// putKey inserts the new key and value into the cache. It returns how
// the internal state was updated and returns a value based on that state.
// It also adds up expiration stat, if the replaced key was expired.
// It also takes in an argument ttl to set a custom expiration time.
func (l *tlruBase[K, V]) putKey(key K, value V, ttl int64) (UpsertState, V) {
	state, val := l.base.putKey(key, ttlValue[V]{
		expiresAt: l.clock.Now() + ttl,
		value:     value,
	})
	if state == Replace {
		if val.expiresAt <= l.clock.Now() {
			l.base.stats.Expirations++
			return AddAfterExpiration, val.value
		}
		return Replace, val.value
	}
	return state, val.value
}

func (l *tlruBase[K, V]) refreshKey(key K, ttl int64) bool {
	curr, ok := l.base.hash[key]
	if !ok {
		return false // not present in cache
	}

	l.base.nodes[curr].value.expiresAt = l.clock.Now() + ttl
	return true
}

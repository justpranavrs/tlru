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

// TTLCore is the implementation of 'LRU' with TTL (Time-To-Live). It
// operates on an internal clock from [lruclock.Clock] and operates with an instance
// of [Core].
type TTLCore[K comparable, V any] struct {
	// core represents the main cache which holds the keys and values.
	core *Core[K, ttlValue[V]]

	// clock is the background timer to ensure fast time loads
	// without halting the LRU operations.
	clock *lruclock.Clock

	// ttl determines the default (time-to-live) duration of an element.
	ttl int64

	// sliding represents if Sliding TTL is enabled.
	// If false Absolute TTL is enabled.
	sliding bool
}

// ttlValue is the wrapper for [Core] with a timestamp and the value.
type ttlValue[V any] struct {
	// expiresAt is the time when the key would get expired.
	expiresAt int64

	// value is the actual data of the cache.
	value V
}

// ttlConfig represents the configuration of [TTLCore]. It should be used with [TTLOption].
type ttlConfig struct {
	// internal clock
	clock *lruclock.Clock

	// sliding TTL
	sliding bool
}

// TTLOption is used to configure [TTLCore] when creating an instance using [NewTTL] constructor.
type TTLOption func(c *ttlConfig)

// NewTTL creates an instance of [TTLCore] using the given capacity and sets
// the default expiration timer based on the argument "ttl".
//
// The ttl value is rounded off in terms of its internal clock ticks.
// Check [lruclock.Clock.Ticks].
//
// It operates on a default clock with 100ms. To customize the
// Clock, refer [WithClock].
//
// Returns an [ErrInvalidCapacity] if the capacity is not in [2, 2147483645].
func NewTTL[K comparable, V any](capacity int, ttl time.Duration, opts ...TTLOption) (*TTLCore[K, V], error) {
	cache, err := New[K, ttlValue[V]](capacity)
	if err != nil {
		return nil, err
	}

	// build the config
	cfg := ttlConfig{}
	for _, opt := range opts { // options
		if opt == nil {
			continue
		}
		opt(&cfg)
	}

	var clock *lruclock.Clock
	if cfg.clock != nil {
		clock = cfg.clock
	} else {
		clock = lruclock.New(100 * time.Millisecond)
		_ = clock.Start()
	}

	return &TTLCore[K, V]{
		core:  cache,
		clock: clock,
		ttl:   clock.Ticks(ttl),
	}, nil
}

// WithClock allows the usage of a custom clock for [TTLCore].
//
// NOTE: Using WithClock on [NewTTL] will not start the clock. Use [lruclock.Clock.Start] to
// initiate the timer.
func WithClock(clock *lruclock.Clock) TTLOption {
	return func(c *ttlConfig) {
		c.clock = clock
	}
}

// WithSliding enables Sliding TTL on the LRU cache.
//
// It will update the timestamp of the key on [TTLCore.Get] and
// [TTLCore.Put] using the TTL provided in [NewTTL].
func WithSliding() TTLOption {
	return func(c *ttlConfig) {
		c.sliding = true
	}
}

// Capacity returns the maximum allocated capacity of the LRU cache.
func (l *TTLCore[K, V]) Capacity() int {
	return l.core.Capacity()
}

// Contains checks whether the key is present in the Cache.
func (l *TTLCore[K, V]) Contains(key K) bool {
	return l.core.Contains(key)
}

// Close safely closes the background clock when TTL is enabled on the cache.
func (l *TTLCore[K, V]) Close() {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	if l.clock != nil {
		l.clock.Stop()
	}
}

// Delete removes the key from the cache and returns the evicted value.
// It returns false if the key was not found in the cache.
func (l *TTLCore[K, V]) Delete(key K) (V, bool) {
	val, ok := l.core.Delete(key)
	return val.value, ok
}

// Flush clears the LRU cache of all its keys and values.
func (l *TTLCore[K, V]) Flush() {
	l.core.Flush()
}

// Get retrieves the cache value using key.
// It returns false if the key is not found or has been expired.
func (l *TTLCore[K, V]) Get(key K) (V, bool) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	val, _, ok := l.getKey(key)
	return val, ok
}

// GetMany allows retrieval of multiple keys at the same time under a single internal lock.
//
// The keys, values and exists array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
//
// The operation modifies values and exists in-place. If a key is not present in the cache or
// has been expired, the corresponding index in exists is set to false
// and leaves the value at that index unchanged.
func (l *TTLCore[K, V]) GetMany(keys []K, values []V, exists []bool) error {
	if (len(keys) != len(values)) || (len(keys) != len(exists)) {
		return ErrInvalidBatchSize
	}
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	for i := range keys {
		val, _, ok := l.getKey(keys[i])
		exists[i] = ok
		if ok {
			values[i] = val
		}
	}
	return nil
}

// Check [TTLCore.Get] on how Get works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *TTLCore[K, V]) GetWithTTL(key K) (V, time.Duration, bool) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	return l.getKey(key)
}

// Peek retrieves the cache value without updating it
// to be the most recently used.
// It returns false if the key is not found or if it has been expired.
func (l *TTLCore[K, V]) Peek(key K) (V, bool) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	val, _, ok := l.peekKey(key)
	return val, ok
}

// Check [TTLCore.Peek] on how Peek works.
// It also returns the remaining TTL in the key if it was found in the cache.
func (l *TTLCore[K, V]) PeekWithTTL(key K) (V, time.Duration, bool) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	return l.peekKey(key)
}

// Put adds a new value to the cache with the given key and assigns a new
// timestamp to the key using the default TTL.
//
// See [TTLCore.Upsert] for detailed information on cache state transitions.
//
// See [TTLCore.PutWithTTL] for adding keys with custom TTL.
func (l *TTLCore[K, V]) Put(key K, value V) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	l.putKey(key, value, l.ttl)
}

// PutMany allows the addition of multiple key-value pairs and also assigns new timestamps
// for the keys, at the same time under a single internal lock.
//
// The keys and values array should be of the same size. If not they are not of same size,
// [ErrInvalidBatchSize] is returned.
func (l *TTLCore[K, V]) PutMany(keys []K, values []V) error {
	if len(keys) != len(values) {
		return ErrInvalidBatchSize
	}
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	for i := range keys {
		l.putKey(keys[i], values[i], l.ttl)
	}
	return nil
}

// PutWithTTL adds a new value to the cache with the provided ttl value.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (l *TTLCore[K, V]) PutWithTTL(key K, value V, ttl time.Duration) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	l.putKey(key, value, l.clock.Ticks(ttl))
}

// Refresh resets the TTL of an existing key using the default ttl.
// It returns false if the key could not be found.
func (l *TTLCore[K, V]) Refresh(key K) bool {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	return l.refreshKey(key, l.ttl)
}

// ResetStats resets the stats of the LRU cache.
func (l *TTLCore[K, V]) ResetStats() {
	l.core.ResetStats()
}

// SetTTL resets the TTL of an existing key using the given ttl in the argument.
// It returns false if the key could not be found.
//
// The ttl value is rounded off in terms of its internal clock ticks.
func (l *TTLCore[K, V]) SetTTL(key K, ttl time.Duration) bool {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	return l.refreshKey(key, l.clock.Ticks(ttl))
}

// Size returns the current size of the LRU cache.
func (l *TTLCore[K, V]) Size() int {
	return l.core.Size()
}

// Stats return the current stats of the LRU cache.
func (l *TTLCore[K, V]) Stats() CoreStats {
	return l.core.Stats()
}

// TTL returns the remaining TTL for the key.
func (l *TTLCore[K, V]) TTL(key K) (time.Duration, bool) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

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
func (l *TTLCore[K, V]) Upsert(key K, value V) (UpsertState, V) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	return l.putKey(key, value, l.ttl)
}

// UpsertWithTTL adds a new value to the cache with the given key and the provided ttl value.
// The ttl value is rounded off in terms of its internal clock ticks.
//
// Check [TTLCore.Upsert] on how Upsert works.
func (l *TTLCore[K, V]) UpsertWithTTL(key K, value V, ttl time.Duration) (UpsertState, V) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	return l.putKey(key, value, l.clock.Ticks(ttl))
}

// expireKey verifies if the timestamp has expired. If it has, it will evict the key.
func (l *TTLCore[K, V]) expireKey(idx int32, val ttlValue[V]) (V, time.Duration, bool) {
	if val.expiresAt < l.clock.Now() {
		l.core.deleteKey(idx)
		l.core.stats.Expirations++

		l.core.stats.Misses++
		return *new(V), 0, false
	}
	if l.sliding {
		l.core.nodes[idx].value.expiresAt = l.clock.Now() + l.ttl
	}
	return val.value, l.clock.Duration() * time.Duration(l.clock.Until(val.expiresAt)), true
}

// getKey retrieves the value and also removes the key if the key has expired.
func (l *TTLCore[K, V]) getKey(key K) (V, time.Duration, bool) {
	curr, ok := l.core.hash[key]
	if !ok {
		return *new(V), 0, false // not present in cache
	}
	return l.expireKey(curr, l.core.getIndex(curr))
}

// peekKey retrieves the value based on the key provided in the argument,
// without ever changing the internal state of the cache unless the key is expired.
func (l *TTLCore[K, V]) peekKey(key K) (V, time.Duration, bool) {
	curr, ok := l.core.hash[key]
	if !ok {
		return *new(V), 0, false // not present in cache
	}
	return l.expireKey(curr, l.core.nodes[curr].value)
}

// putKey inserts the new key and value into the cache. It returns how
// the internal state was updated and returns a value based on that state.
// It also adds up expiration stat, if the replaced key was expired.
// It also takes in an argument ttl to set a custom expiration time.
func (l *TTLCore[K, V]) putKey(key K, value V, ttl int64) (UpsertState, V) {
	state, val := l.core.putKey(key, ttlValue[V]{
		expiresAt: l.clock.Now() + ttl,
		value:     value,
	})
	if state == Replace {
		if val.expiresAt < l.clock.Now() {
			l.core.stats.Expirations++
			return AddAfterExpiration, val.value
		}
		return Replace, val.value
	}
	return state, val.value
}

func (l *TTLCore[K, V]) refreshKey(key K, ttl int64) bool {
	curr, ok := l.core.hash[key]
	if !ok {
		return false // not present in cache
	}

	l.core.nodes[curr].value.expiresAt = l.clock.Now() + ttl
	return true
}

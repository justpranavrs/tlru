// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lrucore

import (
	"time"

	"github.com/justpranavrs/tlru/lruclock"
)

// TTLCore is the implementation of 'LRU' with TTL (Time-To-Live). It
// operates on an internal clock from [lruclock.Clock] and operates with an instance
// of [Core].
type TTLCore[K comparable, V any] struct {
	// core represents the main cache which holds the keys and values.
	core *Core[K, ttlValue[V]]

	// clock is the background timer to ensure fast time loads
	// without halting the LRU operations.
	clock *lruclock.Clock

	// expiresAt determines the default (time-to-live) duration of an element.
	expiresAt int64
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
	clock *lruclock.Clock
}

// TTLOption is used to configure [TTLCore] when creating an instance using [NewTTL] constructor.
type TTLOption func(c *ttlConfig)

// NewTTL creates an instance of [TTLCore] using the given capacity and sets the expiration
// timer based on the argument "expiresAt".
//
// Returns an [ErrInvalidCapacity] if the capacity is not in [2, 2147483645].
func NewTTL[K comparable, V any](capacity int, expiresAt time.Duration, opts ...TTLOption) (*TTLCore[K, V], error) {
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
	duration := 100 * time.Millisecond
	if cfg.clock != nil {
		clock = cfg.clock
		duration = cfg.clock.Duration()
	} else {
		clock = lruclock.New(duration)
		_ = clock.Start()
	}

	return &TTLCore[K, V]{
		core:      cache,
		clock:     clock,
		expiresAt: int64(expiresAt / duration),
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

// Capacity returns the maximum allocated capacity of the LRU cache.
func (l *TTLCore[K, V]) Capacity() int {
	return l.core.Capacity()
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

	return l.getKey(key)
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
// It returns false if the key is not found or if it has been expired.
func (l *TTLCore[K, V]) Peek(key K) (V, bool) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	return l.peekKey(key)
}

// Put adds a new value to the cache with the given key and assigns a new
// timestamp to the key.
// See [Core.Upsert] for detailed information on cache state transitions.
func (l *TTLCore[K, V]) Put(key K, value V) {
	l.core.mu.Lock()
	defer l.core.mu.Unlock()

	l.putKey(key, value)
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
		l.putKey(keys[i], values[i])
	}
	return nil
}

// ResetStats resets the stats of the LRU cache.
func (l *TTLCore[K, V]) ResetStats() {
	l.core.ResetStats()
}

// Size returns the current size of the LRU cache.
func (l *TTLCore[K, V]) Size() int {
	return l.core.Size()
}

// Stats return the current stats of the LRU cache.
func (l *TTLCore[K, V]) Stats() CoreStats {
	return l.core.Stats()
}

// Upsert adds a new value to the cache with the given key and
// also updating their timestamps if replaced.
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

	return l.putKey(key, value)
}

// expireKey verifies if the timestamp has expired. If it has, it will evict the key.
func (l *TTLCore[K, V]) expireKey(idx int32, val ttlValue[V]) (V, bool) {
	if val.expiresAt < l.clock.Now() {
		l.core.deleteKey(idx)
		l.core.stats.Expirations++

		l.core.stats.Misses++
		return *new(V), false
	}
	return val.value, true
}

// getKey retrieves the value and also removes the key if the key has expired.
func (l *TTLCore[K, V]) getKey(key K) (V, bool) {
	curr, ok := l.core.hash[key]
	if !ok {
		return *new(V), false // not present in cache
	}
	return l.expireKey(curr, l.core.getIndex(curr))
}

// peekKey retrieves the value based on the key provided in the argument,
// without ever changing the internal state of the cache unless the key is expired.
func (l *TTLCore[K, V]) peekKey(key K) (V, bool) {
	curr, ok := l.core.hash[key]
	if !ok {
		return *new(V), false // not present in cache
	}
	return l.expireKey(curr, l.core.nodes[curr].value)
}

// putKey inserts the new key and value into the cache. It returns how
// the internal state was updated and returns a value based on that state.
// It also adds up expiration stat, if the replaced key was expired
func (l *TTLCore[K, V]) putKey(key K, value V) (UpsertState, V) {
	state, val := l.core.putKey(key, ttlValue[V]{
		expiresAt: l.clock.Now() + l.expiresAt,
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

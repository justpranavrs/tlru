// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import "time"

// Shard defines the most basic blueprint of a 'Least Recently Used' cache.
//
// As Items are Added to the Cache, The 'Least Recently Used' key
// is evicted from the Cache.
//
// K represents the type of the key, whereas V represents the type of the Value
// in the cache
type Shard[K comparable, V any] interface {
	// Capacity returns the maximum allocated capacity of the LRU cache.
	Capacity() int

	// Contains checks whether the key is present in the Cache.
	Contains(key K) bool

	// Delete removes the key from the cache and returns the evicted value.
	// It returns false if the key was not found in the cache.
	Delete(key K) (V, bool)

	// Flush clears the LRU cache of all its keys and values.
	Flush()

	// Get retrieves the cache value using key.
	// It returns false if the key is not found.
	Get(key K) (V, bool)

	// GetMany allows retrieval of multiple keys at the same time.
	GetMany(keys []K, values []V, exists []bool) error

	// Peek retrieves the cache value without updating it
	// to be the most recently used.
	// It returns false if the key is not found.
	Peek(key K) (V, bool)

	// Put adds a new value to the cache with the given key.
	Put(key K, value V)

	// PutMany adds multiple key-value pairs at the same time.
	PutMany(keys []K, values []V) error

	// ResetStats resets the stats of the LRU cache.
	ResetStats()

	// Size returns the current size of the LRU cache.
	Size() int

	// Stats return the current stats of the LRU cache.
	Stats() Stats

	// Upsert adds a new value to the cache with the given key.
	// It returns a value based on how the internal state of the cache changed.
	Upsert(key K, value V) (UpsertState, V)
}

// TTLShard defines the blueprint of a 'Time-Aware Least Recently Used'.
// It implements [Shard].
type TTLShard[K comparable, V any] interface {
	Shard[K, V]

	// Close safely closes the background clock when TTL is enabled on the cache.
	Close()

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

// Stats represents the metrics of a [Shard] instance.
type Stats struct {
	// Hits is the number of successful cache lookups from [Shard.Get] and [Shard.GetMany].
	Hits int

	// Misses is the number of failed cache lookups from [Shard.Get] and [Shard.GetMany].
	Misses int

	// Evictions is the number of keys removed from cache during [Shard.Put] and [Shard.PutMany].
	Evictions int

	// Expirations is the number of expirations triggered due to TTL (Time-To-Live).
	Expirations int
}

// UpsertState represents the value returned by Upsert operation.
type UpsertState uint8

const (
	// AddAfterExpiration is triggered when a new key was added after an expiration
	// due to TTL (Time-To-Live).
	AddAfterExpiration UpsertState = iota

	// AddNoEvict is triggered when a new key was added without eviction.
	AddNoEvict

	// AddOnEvict is triggered when a new key was added after an eviction.
	AddOnEvict

	// Replace is triggered when an older key's value was replaced.
	Replace
)

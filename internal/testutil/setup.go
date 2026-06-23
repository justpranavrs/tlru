// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testutil

import (
	"math"

	"github.com/justpranavrs/tlru/lrucore"
)

// methods enum
const (
	opInit = iota
	opCapacity
	opFlush
	opGet
	opGetMany
	opPeek
	opPut
	opPutMany
	opSize
	opUpsert
)

// CacheTest is a testing interface for the LRU instances.
// For more details, refer [tlru.Cache]
type CacheTest[K comparable, V any] interface {
	Capacity() int
	Flush()
	Get(key K) (V, bool)
	Peek(key K) (V, bool)
	Put(key K, value V)
	Upsert(key K, value V) lrucore.UpsertState
	Size() int
}

// actions are defined for the fuzz test.
// flush is removed because fuzz tests are for accuracy.
var actions = []int{
	opGet, opGet, opGet, opGet, opGet, opGet, opPeek, opPeek,
	opPut, opPut, opPut, opPut, opUpsert, opUpsert, opUpsert, opSize,
}

const fuzzBytes int = 2
const fuzzKeys int = 4096

// testCacheOp defines the structure of a unit test data.
type testCacheOp struct {
	// input
	method int
	key    int
	value  User

	capacity int // for opInit only

	// output
	expectedValue  User
	expectedNumber int
	expectedBool   bool
}

// finds the evict index
func evictKey(tick []int, shard uint32, mux func(int) uint32) int {
	evictIdx := 0
	evictTick := math.MaxInt32

	for idx := range tick {
		sh := mux(idx) // checks if it is the current shard.
		if sh != shard {
			continue
		}

		// if this is the least tick not equal to -1, no two keys can have same tick
		if tick[idx] != -1 && tick[idx] < evictTick {
			evictIdx = idx
			evictTick = tick[idx]
		}
	}
	return evictIdx
}

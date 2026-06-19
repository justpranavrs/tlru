// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testutil

import "math"

// methods enum
const (
	opInit = iota
	opCapacity
	opContains
	opFlush
	opGet
	opPeek
	opPut
	opSize
)

// actions are defined for the fuzz test.
// flush is removed because fuzz tests are for accuracy.
var actions = []int{
	opContains, opContains, opGet, opGet, opGet, opGet, opGet, opGet,
	opContains, opPeek, opPeek, opPut, opPut, opPut, opPut, opSize,
}

// CacheOp contains of the method (put or get) with key and value.
type CacheOp struct {
	method int
	key    int
	value  User
}

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
func evictKey(tick []int, shard uint32, keys int, mux func(int) uint32) int {
	evictIdx := 0
	evictTick := math.MaxInt32 - 1

	for idx := range keys {
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

func TestMux(mask uint32) func(int) uint32 {
	return func(key int) uint32 {
		var hash uint32 = 0
		for i := 0; i < 8; i++ {
			hash ^= uint32(key & 255)
			key >>= 8
			hash *= 16777619
		}
		return (hash & mask)
	}
}

// BasicTestData represents the basic unit test's data.
var BasicTestData = []testCacheOp{
	{method: opInit, capacity: 256},
	{method: opPut, key: 1, value: User{Name: "justpranavrs", Email: "iliketlru@gmail.com"}},
	{method: opGet, key: 1, expectedBool: true, expectedValue: User{Name: "justpranavrs", Email: "iliketlru@gmail.com"}},
	{method: opSize, expectedNumber: 1},
	{method: opPut, key: 3, value: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	{method: opPut, key: 4, value: User{Name: "golang-tlru", Email: "tlrupeace@gmail.com"}},
	{method: opSize, expectedNumber: 3},
	{method: opCapacity, expectedNumber: 256},
	{method: opGet, key: 2, expectedBool: false},
	{method: opPut, key: 2, value: User{Name: "just-golang", Email: "justtlru@gmail.com"}},
	{method: opGet, key: 3, expectedBool: true, expectedValue: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	{method: opContains, key: 5, expectedBool: false},
	{method: opGet, key: 1, expectedBool: true, expectedValue: User{Name: "justpranavrs", Email: "iliketlru@gmail.com"}},
	{method: opSize, expectedNumber: 4},
	{method: opPeek, key: 4, expectedBool: true, expectedValue: User{Name: "golang-tlru", Email: "tlrupeace@gmail.com"}},
	{method: opFlush},
	{method: opGet, key: 2, expectedBool: false},
	{method: opSize, expectedNumber: 0},
	{method: opCapacity, expectedNumber: 256},
	{method: opPut, key: 7, value: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	{method: opSize, expectedNumber: 1},
}

// AdvancedTestData represents the advanced unit test's data.
var AdvancedTestData = []testCacheOp{
	{method: opInit, capacity: 3},
	{method: opPut, key: 1, value: User{Name: "justpranavrs", Email: "iliketlru@gmail.com"}},
	// 1
	{method: opPut, key: 3, value: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	// 1 3
	{method: opGet, key: 1, expectedBool: true, expectedValue: User{Name: "justpranavrs", Email: "iliketlru@gmail.com"}},
	// 3 1
	{method: opSize, expectedNumber: 2},
	{method: opPut, key: 4, value: User{Name: "golang-tlru", Email: "tlrupeace@gmail.com"}},
	// 3 1 4
	{method: opSize, expectedNumber: 3},
	{method: opCapacity, expectedNumber: 3},
	{method: opGet, key: 2, expectedBool: false},
	{method: opPut, key: 2, value: User{Name: "just-golang", Email: "justtlru@gmail.com"}},
	// 1 4 2
	{method: opGet, key: 3, expectedBool: false},
	{method: opPut, key: 1, value: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	// 4 2 1
	{method: opSize, expectedNumber: 3},
	{method: opContains, key: 4, expectedBool: true},
	{method: opPeek, key: 4, expectedBool: true, expectedValue: User{Name: "golang-tlru", Email: "tlrupeace@gmail.com"}},
	{method: opCapacity, expectedNumber: 3},
	{method: opContains, key: 2, expectedBool: true},
	{method: opPut, key: 7, value: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	// 2 1 7
	{method: opContains, key: 4, expectedBool: false},
	{method: opSize, expectedNumber: 3},
	{method: opContains, key: 1, expectedBool: true},
	{method: opContains, key: 2, expectedBool: true},
	{method: opGet, key: 4, expectedBool: false},
	{method: opGet, key: 3, expectedBool: false},
	{method: opGet, key: 1, expectedBool: true, expectedValue: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	// 2 7 1
	{method: opPut, key: 8, value: User{Name: "golang-tlru", Email: "tlrupeace@gmail.com"}},
	// 7 1 8
	{method: opContains, key: 7, expectedBool: true},
	{method: opContains, key: 2, expectedBool: false},
	{method: opSize, expectedNumber: 3},
	{method: opFlush},
	{method: opSize, expectedNumber: 0},
	{method: opCapacity, expectedNumber: 3},
}

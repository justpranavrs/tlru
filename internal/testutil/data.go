// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testutil

import "github.com/justpranavrs/tlru/lrucore"

// BasicCoreData represents the basic unit test's data for [lrucore.Core].
var BasicCoreData = []testCacheOp{
	// 0
	{method: opInit, capacity: 256},
	{method: opPut, key: 1, value: User{Name: "justpranavrs", Email: "iliketlru@gmail.com"}},
	{method: opGet, key: 1, expectedBool: true, expectedValue: User{Name: "justpranavrs", Email: "iliketlru@gmail.com"}},
	{method: opSize, expectedNumber: 1},
	{method: opPut, key: 3, value: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},

	// 5
	{method: opPut, key: 4, value: User{Name: "justpranavrs", Email: "iliketlru@gmail.com"}},
	{method: opSize, expectedNumber: 3},
	{method: opCapacity, expectedNumber: 256},
	{method: opGet, key: 2, expectedBool: false},
	{method: opPut, key: 2, value: User{Name: "just-golang", Email: "justtlru@gmail.com"}},

	// 10
	{method: opGet, key: 3, expectedBool: true, expectedValue: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	{method: opPeek, key: 5, expectedBool: false},
	{method: opGet, key: 1, expectedBool: true, expectedValue: User{Name: "justpranavrs", Email: "iliketlru@gmail.com"}},
	{method: opSize, expectedNumber: 4},
	{method: opUpsert, key: 7, value: User{Name: "golang-is-awesome", Email: "goisawesum@gmail.com"},
		expectedNumber: int(lrucore.AddNoEvict),
	},
	{method: opPeek, key: 2, expectedBool: true, expectedValue:  User{Name: "just-golang", Email: "justtlru@gmail.com"}},
	{method: opFlush},
	{method: opGet, key: 2, expectedBool: false},
	{method: opSize, expectedNumber: 0},

	// 21
	{method: opCapacity, expectedNumber: 256},
	{method: opPut, key: 7, value: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	{method: opSize, expectedNumber: 1},
}

// AdvancedCoreData represents the advanced unit test's data for [tlru.LRU].
var AdvancedCoreData = []testCacheOp{
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
	{method: opPeek, key: 4, expectedBool: true, expectedValue: User{Name: "golang-tlru", Email: "tlrupeace@gmail.com"}},
	{method: opCapacity, expectedNumber: 3},
	{method: opPeek, key: 2, expectedBool: true, expectedValue: User{Name: "just-golang", Email: "justtlru@gmail.com"}},
	{method: opPut, key: 7, value: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	// 2 1 7
	{method: opPeek, key: 4, expectedBool: false},
	{method: opSize, expectedNumber: 3},
	{method: opPeek, key: 1, expectedBool: true, expectedValue: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	{method: opPeek, key: 2, expectedBool: true, expectedValue: User{Name: "just-golang", Email: "justtlru@gmail.com"}},
	{method: opGet, key: 4, expectedBool: false},
	{method: opGet, key: 3, expectedBool: false},
	{method: opGet, key: 1, expectedBool: true, expectedValue: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	// 2 7 1
	{method: opPut, key: 8, value: User{Name: "golang-tlru", Email: "tlrupeace@gmail.com"}},
	// 7 1 8
	{method: opPeek, key: 7, expectedBool: true, expectedValue: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	{method: opPeek, key: 2, expectedBool: false},
	{method: opSize, expectedNumber: 3},
	{method: opFlush},
	{method: opSize, expectedNumber: 0},
	{method: opCapacity, expectedNumber: 3},
}

// BasicLRUData represents the basic unit test's data for [tlru.LRU].
var BasicLRUData = []testCacheOp{
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
	{method: opPeek, key: 5, expectedBool: false},
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

// AdvancedLRUData represents the advanced unit test's data for [tlru.LRU].
var AdvancedLRUData = []testCacheOp{
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
	{method: opPeek, key: 4, expectedBool: true, expectedValue: User{Name: "golang-tlru", Email: "tlrupeace@gmail.com"}},
	{method: opCapacity, expectedNumber: 3},
	{method: opPeek, key: 2, expectedBool: true, expectedValue: User{Name: "just-golang", Email: "justtlru@gmail.com"}},
	{method: opPut, key: 7, value: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	// 2 1 7
	{method: opPeek, key: 4, expectedBool: false},
	{method: opSize, expectedNumber: 3},
	{method: opPeek, key: 1, expectedBool: true, expectedValue: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	{method: opPeek, key: 2, expectedBool: true, expectedValue: User{Name: "just-golang", Email: "justtlru@gmail.com"}},
	{method: opGet, key: 4, expectedBool: false},
	{method: opGet, key: 3, expectedBool: false},
	{method: opGet, key: 1, expectedBool: true, expectedValue: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	// 2 7 1
	{method: opPut, key: 8, value: User{Name: "golang-tlru", Email: "tlrupeace@gmail.com"}},
	// 7 1 8
	{method: opPeek, key: 7, expectedBool: true, expectedValue: User{Name: "tlru", Email: "tlruiscool@gmail.com"}},
	{method: opPeek, key: 2, expectedBool: false},
	{method: opSize, expectedNumber: 3},
	{method: opFlush},
	{method: opSize, expectedNumber: 0},
	{method: opCapacity, expectedNumber: 3},
}

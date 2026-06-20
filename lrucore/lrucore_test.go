// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lrucore_test

import (
	"testing"

	"github.com/justpranavrs/tlru/internal/testutil"
	"github.com/justpranavrs/tlru/lrucore"
)

// TestLRUCore runs a basic and advanced unit tests for the core LRU instance.
func TestLRUCore(t *testing.T) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := lrucore.New[int, testutil.User](capacity)
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicCoreData, init)
	testutil.TestCache(t, testutil.AdvancedCoreData, init)
}

// TestRaceLRUCore_Int runs a concurrency test for the sharded LRU instance with int keys.
func TestRaceLRUCore_Int(t *testing.T) {
	cache, err := lrucore.New[int, testutil.User](512)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 24
	numWorkers := 256

	testutil.TestRaceCache(t, cache, keys, numOps, numWorkers, func(c testutil.CacheOp) int {
		return c.Key
	})
}

// TestRaceLRUCore_Int runs a concurrency test for the sharded LRU instance with int32 keys.
func TestRaceLRUCore_Int32(t *testing.T) {
	cacheInt32, err := lrucore.New[int32, testutil.User](512)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 24
	numWorkers := 256

	testutil.TestRaceCache(t, cacheInt32, keys, numOps, numWorkers, func(c testutil.CacheOp) int32 {
		return int32(c.Key)
	})
}

// TestRaceLRUCore_Int runs a concurrency test for the sharded LRU instance with uint keys.
func TestRaceLRUCore_Uint(t *testing.T) {
	cacheUint, err := lrucore.New[uint, testutil.User](512)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 24
	numWorkers := 256

	testutil.TestRaceCache(t, cacheUint, keys, numOps, numWorkers, func(c testutil.CacheOp) uint {
		return uint(c.Key)
	})
}

// TestRaceLRUCore_Int runs a concurrency test for the sharded LRU instance with string keys.
func TestRaceLRUCore_String(t *testing.T) {
	cacheStr, err := lrucore.New[uint, testutil.User](512)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 24
	numWorkers := 256

	testutil.TestRaceCache(t, cacheStr, keys, numOps, numWorkers, func(c testutil.CacheOp) uint {
		return uint(c.Key)
	})
}

// FuzzLRUCore runs a fuzz test for the core LRU instance.
func FuzzLRUCore(f *testing.F) {
	cache, err := lrucore.New[int, testutil.User](512)
	if err != nil {
		f.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.FuzzCache(f, cache, 8192, 2, 512, 1)
}

// BenchmarkLRUCore runs a benchmark test for the core LRU instance.
func BenchmarkLRUCore(b *testing.B) {
	cache, err := lrucore.New[int, testutil.User](512)
	if err != nil {
		b.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 24

	zipFData := testutil.GenerateZipfData(keys, numOps)
	randomData := testutil.GenerateRandomData(keys, numOps)
	testutil.BenchmarkCache(b, cache, "Zipf", 512, numOps, zipFData)
	testutil.BenchmarkCache(b, cache, "Random", 512, numOps, randomData)
}

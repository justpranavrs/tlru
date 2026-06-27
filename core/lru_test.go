// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core_test

import (
	"testing"

	"github.com/justpranavrs/tlru/core"
	"github.com/justpranavrs/tlru/internal/testutil"
)

// TestLRU runs a basic and advanced unit tests for the [LRU] instance.
func TestLRU(t *testing.T) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := core.New[int, testutil.User](capacity)
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicLRUData, init)
	testutil.TestCache(t, testutil.AdvancedLRUData, init)
}

// TestRaceLRU_Int runs a concurrency test for the [LRU] instance with int keys.
func TestRaceLRU_Int(t *testing.T) {
	cache, err := core.New[int, testutil.User](2048)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 20
	numWorkers := 256

	testutil.TestRaceCache(t, cache, keys, numOps, numWorkers, func(c testutil.CacheOp) int {
		return c.Key
	})
}

// TestRaceLRU_Int runs a concurrency test for the [LRU] instance with int32 keys.
func TestRaceLRU_Int32(t *testing.T) {
	cacheInt32, err := core.New[int32, testutil.User](2048)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 20
	numWorkers := 256

	testutil.TestRaceCache(t, cacheInt32, keys, numOps, numWorkers, func(c testutil.CacheOp) int32 {
		return int32(c.Key)
	})
}

// TestRaceLRU_Int runs a concurrency test for the [LRU] instance with uint keys.
func TestRaceLRU_Uint(t *testing.T) {
	cacheUint, err := core.New[uint, testutil.User](2048)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 20
	numWorkers := 256

	testutil.TestRaceCache(t, cacheUint, keys, numOps, numWorkers, func(c testutil.CacheOp) uint {
		return uint(c.Key)
	})
}

// TestRaceLRU_Int runs a concurrency test for the [LRU] instance with string keys.
func TestRaceLRU_String(t *testing.T) {
	cacheStr, err := core.New[string, testutil.User](2048)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 20
	numWorkers := 256

	testutil.TestRaceCache(t, cacheStr, keys, numOps, numWorkers, func(c testutil.CacheOp) string {
		return c.Value.Name
	})
}

// FuzzLRU runs a fuzz test for the [LRU] instance.
func FuzzLRU(f *testing.F) {
	cache, err := core.New[int, testutil.User](512)
	if err != nil {
		f.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.FuzzCache(f, cache, func(i int) uint32 { return 0 }, 8192, 512, 1536, 1)
}

// BenchmarkLRU runs a benchmark test for the [LRU] instance.
func BenchmarkLRU(b *testing.B) {
	cache, err := core.New[int, testutil.User](2048)
	if err != nil {
		b.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 65536
	numOps := 1 << 20

	zipFData := testutil.GenerateZipfData(keys, numOps)
	randomData := testutil.GenerateRandomData(keys, numOps)
	testutil.BenchmarkCache(b, cache, "Zipf", 2048, numOps, zipFData)
	testutil.BenchmarkCache(b, cache, "Random", 2048, numOps, randomData)
}

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core_test

import (
	"testing"

	"github.com/justpranavrs/tlru/core"
	"github.com/justpranavrs/tlru/internal/testutil"
)

// TestSLRU runs a basic and advanced unit tests for the [SLRU] instance.
func TestSLRU(t *testing.T) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := core.NewSLRU[int, testutil.User](capacity, 20)
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicLRUData, init)
}

// TestRaceSLRU_Int runs a concurrency test for the [SLRU] instance with int keys.
func TestRaceSLRU_Int(t *testing.T) {
	cache, err := core.NewSLRU[int, testutil.User](2048, 20)
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

// TestRaceSLRU_Int runs a concurrency test for the [SLRU] instance with int32 keys.
func TestRaceSLRU_Int32(t *testing.T) {
	cacheInt32, err := core.NewSLRU[int32, testutil.User](2048, 20)
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

// TestRaceSLRU_Int runs a concurrency test for the [SLRU] instance with uint keys.
func TestRaceSLRU_Uint(t *testing.T) {
	cacheUint, err := core.NewSLRU[uint, testutil.User](2048, 20)
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

// TestRaceSLRU_Int runs a concurrency test for the [SLRU] instance with string keys.
func TestRaceSLRU_String(t *testing.T) {
	cacheStr, err := core.NewSLRU[string, testutil.User](2048, 20)
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

// BenchmarkSLRU runs a benchmark test for the [SLRU] instance.
func BenchmarkSLRU(b *testing.B) {
	cache, err := core.NewSLRU[int, testutil.User](2048, 20)
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

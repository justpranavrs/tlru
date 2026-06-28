// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core_test

import (
	"testing"
	"time"

	"github.com/justpranavrs/tlru/core"
	"github.com/justpranavrs/tlru/internal/testutil"
)

// TestTLRU runs a basic and advanced unit tests for the [TLRU] instance.
func TestTLRU(t *testing.T) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := core.NewWithTTL[int, testutil.User](capacity, time.Hour)
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicLRUData, init)
	testutil.TestCache(t, testutil.AdvancedLRUData, init)
}

// TestRaceTLRU_Int runs a concurrency test for the [TLRU] instance with int keys.
func TestRaceTLRU_Int(t *testing.T) {
	cache, err := core.NewWithTTL[int, testutil.User](2048, time.Hour)
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

// TestRaceTLRU_Int runs a concurrency test for the [TLRU] instance with int32 keys.
func TestRaceTLRU_Int32(t *testing.T) {
	cacheInt32, err := core.NewWithTTL[int32, testutil.User](2048, time.Hour)
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

// TestRaceTLRU_Int runs a concurrency test for the [TLRU] instance with uint keys.
func TestRaceTLRU_Uint(t *testing.T) {
	cacheUint, err := core.NewWithTTL[uint, testutil.User](2048, time.Hour)
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

// TestRaceTLRU_Int runs a concurrency test for the [TLRU] instance with string keys.
func TestRaceTLRU_String(t *testing.T) {
	cacheStr, err := core.NewWithTTL[string, testutil.User](2048, time.Hour)
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

// FuzzTLRU runs a fuzz test for the [TLRU] instance.
func FuzzTLRU(f *testing.F) {
	cache, err := core.NewWithTTL[int, testutil.User](512, time.Hour)
	if err != nil {
		f.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.FuzzCache(f, cache, func(i int) uint32 { return 0 }, 8192, 512, 1536, 1)
}

// BenchmarkTLRU runs a benchmark test for the [TLRU] instance.
func BenchmarkTLRU(b *testing.B) {
	cache, err := core.NewWithTTL[int, testutil.User](2048, time.Hour)
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

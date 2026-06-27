// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru_test

import (
	"testing"

	"github.com/justpranavrs/tlru"
	"github.com/justpranavrs/tlru/internal/testutil"
	"github.com/justpranavrs/tlru/mux"
)

// TestPoolLRU runs a basic unit test for the sharded LRU instance.
func TestPoolLRU(t *testing.T) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := tlru.New[int, testutil.User](capacity)
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicLRUData, init)
}

// TestRacePoolLRU_Int runs a concurrency test for the sharded LRU instance with int keys.
func TestRacePoolLRU_Int(t *testing.T) {
	cache, err := tlru.New[int, testutil.User](512)
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

// TestRacePoolLRU_Int runs a concurrency test for the sharded LRU instance with int32 keys.
func TestRacePoolLRU_Int32(t *testing.T) {
	cacheInt32, err := tlru.New[int32, testutil.User](512)
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

// TestRacePoolLRU_Int runs a concurrency test for the sharded LRU instance with uint keys.
func TestRacePoolLRU_Uint(t *testing.T) {
	cacheUint, err := tlru.New[uint, testutil.User](512)
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

// TestRacePoolLRU_Int runs a concurrency test for the sharded LRU instance with string keys.
func TestRacePoolLRU_String(t *testing.T) {
	cacheStr, err := tlru.New[string, testutil.User](512)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 24
	numWorkers := 256

	testutil.TestRaceCache(t, cacheStr, keys, numOps, numWorkers, func(c testutil.CacheOp) string {
		return c.Value.Email
	})
}

// FuzzPoolLRU runs a fuzz test for the sharded LRU instance.
func FuzzPoolLRU(f *testing.F) {
	hasher := mux.NewMH32[int](tlru.DefaultShards)
	cache, err := tlru.New[int, testutil.User](512, tlru.WithMux(hasher))
	if err != nil {
		f.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.FuzzCache(f, cache, hasher, 8192, 512, tlru.DefaultShards)
}

// BenchmarkPoolLRUWith64 runs a benchmark test for the sharded LRU instance
// with 64 sharded [core.LRU] instances.
func BenchmarkPoolLRUWith64(b *testing.B) {
	cache, err := tlru.New[int, testutil.User](512, tlru.WithShards(64))
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

// BenchmarkPoolLRU runs a benchmark test for the sharded LRU instance.
func BenchmarkPoolLRU(b *testing.B) {
	cache, err := tlru.New[int, testutil.User](512)
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

// BenchmarkPoolLRUWith256 runs a benchmark test for the sharded LRU instance
// with 256 sharded [core.LRU] instances.
func BenchmarkPoolLRUWith256(b *testing.B) {
	cache, err := tlru.New[int, testutil.User](512, tlru.WithShards(256))
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

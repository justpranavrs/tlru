// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru_test

import (
	"testing"

	"github.com/justpranavrs/tlru"
	"github.com/justpranavrs/tlru/internal/testutil"
)

// TestPoolSLRU runs a basic unit test for the sharded LRU instance.
func TestPoolSLRU(t *testing.T) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := tlru.NewSLRU[int, testutil.User](capacity, 10, tlru.WithShards(4))
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicLRUData, init)
}

// TestRacePoolSLRU_Int runs a concurrency test for the sharded LRU instance with int keys.
func TestRacePoolSLRU_Int(t *testing.T) {
	cache, err := tlru.NewSLRU[int, testutil.User](4096, 20)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 65536
	numOps := 1 << 20
	numWorkers := 256

	testutil.TestRaceCache(t, cache, keys, numOps, numWorkers, func(c testutil.CacheOp) int {
		return c.Key
	})
}

// TestRacePoolSLRU_Int runs a concurrency test for the sharded LRU instance with int32 keys.
func TestRacePoolSLRU_Int32(t *testing.T) {
	cacheInt32, err := tlru.NewSLRU[int32, testutil.User](4096, 20)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 65536
	numOps := 1 << 20
	numWorkers := 256

	testutil.TestRaceCache(t, cacheInt32, keys, numOps, numWorkers, func(c testutil.CacheOp) int32 {
		return int32(c.Key)
	})
}

// TestRacePoolSLRU_Int runs a concurrency test for the sharded LRU instance with uint keys.
func TestRacePoolSLRU_Uint(t *testing.T) {
	cacheUint, err := tlru.NewSLRU[uint, testutil.User](4096, 20)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 65536
	numOps := 1 << 20
	numWorkers := 256

	testutil.TestRaceCache(t, cacheUint, keys, numOps, numWorkers, func(c testutil.CacheOp) uint {
		return uint(c.Key)
	})
}

// TestRacePoolSLRU_Int runs a concurrency test for the sharded LRU instance with string keys.
func TestRacePoolSLRU_String(t *testing.T) {
	cacheStr, err := tlru.NewSLRU[string, testutil.User](4096, 20)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 65536
	numOps := 1 << 20
	numWorkers := 256

	testutil.TestRaceCache(t, cacheStr, keys, numOps, numWorkers, func(c testutil.CacheOp) string {
		return c.Value.Email
	})
}

// BenchmarkPoolSLRUWith64 runs a benchmark test for the sharded LRU instance
// with 64 sharded [core.LRU] instances.
func BenchmarkPoolSLRUWith64(b *testing.B) {
	cache, err := tlru.NewSLRU[int, testutil.User](16384, 20, tlru.WithShards(64))
	if err != nil {
		b.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 524288
	numOps := 1 << 20

	zipFData := testutil.GenerateZipfData(keys, numOps)
	randomData := testutil.GenerateRandomData(keys, numOps)
	testutil.BenchmarkCache(b, cache, "Zipf", 16384, numOps, zipFData)
	testutil.BenchmarkCache(b, cache, "Random", 16384, numOps, randomData)
}

// BenchmarkPoolSLRU runs a benchmark test for the sharded LRU instance.
func BenchmarkPoolSLRU(b *testing.B) {
	cache, err := tlru.NewSLRU[int, testutil.User](16384, 20)
	if err != nil {
		b.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 524288
	numOps := 1 << 20

	zipFData := testutil.GenerateZipfData(keys, numOps)
	randomData := testutil.GenerateRandomData(keys, numOps)
	testutil.BenchmarkCache(b, cache, "Zipf", 16384, numOps, zipFData)
	testutil.BenchmarkCache(b, cache, "Random", 16384, numOps, randomData)
}

// BenchmarkPoolSLRUWith256 runs a benchmark test for the sharded LRU instance
// with 256 sharded [core.LRU] instances.
func BenchmarkPoolSLRUWith256(b *testing.B) {
	cache, err := tlru.NewSLRU[int, testutil.User](16384, 20, tlru.WithShards(256))
	if err != nil {
		b.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 524288
	numOps := 1 << 20

	zipFData := testutil.GenerateZipfData(keys, numOps)
	randomData := testutil.GenerateRandomData(keys, numOps)
	testutil.BenchmarkCache(b, cache, "Zipf", 16384, numOps, zipFData)
	testutil.BenchmarkCache(b, cache, "Random", 16384, numOps, randomData)
}

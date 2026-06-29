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
		cache, err := tlru.NewSegmented[int, testutil.User](capacity, 10, tlru.WithShards(4))
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicLRUData, init)
}

// TestRaceSLRU runs [raceSegmented] for different types of keys.
func TestRaceSLRU(t *testing.T) {
	raceSegmented[int32](t, "int32")
	raceSegmented[int](t, "int")
	raceSegmented[uint](t, "uint")
	raceSegmented[string](t, "string")
}

// raceSegmented runs a concurrency test for the [PoolSLRU] instance with keys.
func raceSegmented[K comparable](t *testing.T, key string) {
	cache, err := tlru.NewSegmented[K, testutil.User](2048, 30)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.RunRace(t, key, cache)
}

// BenchmarkPoolSLRUWith64 runs a benchmark test for the sharded LRU instance
// with 64 sharded [core.LRU] instances.
func BenchmarkPoolSLRUWith64(b *testing.B) {
	cache, err := tlru.NewSegmented[int, testutil.User](16384, 20, tlru.WithShards(64))
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
	cache, err := tlru.NewSegmented[int, testutil.User](16384, 20)
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
	cache, err := tlru.NewSegmented[int, testutil.User](16384, 20, tlru.WithShards(256))
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

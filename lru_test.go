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

// TestRaceLRU runs [racePool] for different types of keys.
func TestRaceLRU(t *testing.T) {
	racePool[int32](t, "int32")
	racePool[int](t, "int")
	racePool[uint](t, "uint")
	racePool[string](t, "string")
}

// racePool runs a concurrency test for the [PoolLRU] instance with keys.
func racePool[K comparable](t *testing.T, key string) {
	cache, err := tlru.New[K, testutil.User](2048)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.RunRace(t, key, cache)
}

// FuzzPoolLRU runs a fuzz test for the sharded LRU instance.
func FuzzPoolLRU(f *testing.F) {
	hasher := mux.NewMH32[int](tlru.DefaultShards)
	cache, err := tlru.New[int, testutil.User](512, tlru.WithMux(hasher))
	if err != nil {
		f.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.FuzzCache(f, cache, hasher, 8192, 512, 1536, tlru.DefaultShards)
}

// BenchmarkPoolLRUWith64 runs a benchmark test for the sharded LRU instance
// with 64 sharded [core.LRU] instances.
func BenchmarkPoolLRUWith64(b *testing.B) {
	cache, err := tlru.New[int, testutil.User](16384, tlru.WithShards(64))
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

// BenchmarkPoolLRU runs a benchmark test for the sharded LRU instance.
func BenchmarkPoolLRU(b *testing.B) {
	cache, err := tlru.New[int, testutil.User](16384)
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

// BenchmarkPoolLRUWith256 runs a benchmark test for the sharded LRU instance
// with 256 sharded [core.LRU] instances.
func BenchmarkPoolLRUWith256(b *testing.B) {
	cache, err := tlru.New[int, testutil.User](16384, tlru.WithShards(256))
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

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru_test

import (
	"testing"

	"github.com/justpranavrs/tlru"
	"github.com/justpranavrs/tlru/internal/testutil"
)

// TestLRU runs a basic unit test for the sharded LRU instance.
func TestLRU(t *testing.T) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := tlru.New[int, testutil.User](capacity)
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicTestData, init)
}

// TestRaceLRU runs a concurrency test for the sharded LRU instance.
func TestRaceLRU(t *testing.T) {
	cache, err := tlru.New[int, testutil.User](512)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}

	keys := 16384
	numOps := 1 << 24

	testutil.TestRaceCache(t, cache, keys, numOps, 64)
}

// FuzzLRU runs a fuzz test for the sharded LRU instance.
func FuzzLRU(f *testing.F) {
	cache, err := tlru.New[int, testutil.User](512,
		tlru.WithMux(testutil.TestMux(uint32(tlru.DefaultShards)-1)),
	)
	if err != nil {
		f.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.FuzzCache(f, cache, 8192, 2, 512, tlru.DefaultShards)
}

// BenchmarkLRUWith64 runs a benchmark test for the sharded LRU instance
// with 64 sharded [lrucore.LRUCore] instances.
func BenchmarkLRUWith64(b *testing.B) {
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

// BenchmarkLRU runs a benchmark test for the sharded LRU instance.
func BenchmarkLRU(b *testing.B) {
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

// BenchmarkLRUWith256 runs a benchmark test for the sharded LRU instance
// with 256 sharded [lrucore.LRUCore] instances.
func BenchmarkLRUWith256(b *testing.B) {
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

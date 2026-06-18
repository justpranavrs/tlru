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

// FuzzLRU runs a fuzz test for the sharded LRU instance.
func FuzzLRU(f *testing.F) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := tlru.New[int, testutil.User](capacity)
		if err != nil {
			f.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.FuzzCache(f, init, 512, 8192, 2)
}

// BenchmarkLRUWith64 runs a benchmark test for the sharded LRU instance
// with 64 sharded [lrucore.LRUCore] instances.
func BenchmarkLRUWith64(b *testing.B) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := tlru.New[int, testutil.User](capacity, tlru.WithShards(64))
		if err != nil {
			b.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}

	keys := 16384
	numOps := 2 << 24

	zipFData := testutil.GenerateZipfData(keys, numOps)
	randomData := testutil.GenerateRandomData(keys, numOps)
	testutil.BenchmarkCache(b, init, "Zipf", 512, numOps, zipFData)
	testutil.BenchmarkCache(b, init, "Random", 512, numOps, randomData)
}

// BenchmarkLRU runs a benchmark test for the sharded LRU instance.
func BenchmarkLRU(b *testing.B) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := tlru.New[int, testutil.User](capacity)
		if err != nil {
			b.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}

	keys := 16384
	numOps := 2 << 24

	zipFData := testutil.GenerateZipfData(keys, numOps)
	randomData := testutil.GenerateRandomData(keys, numOps)
	testutil.BenchmarkCache(b, init, "Zipf", 512, numOps, zipFData)
	testutil.BenchmarkCache(b, init, "Random", 512, numOps, randomData)
}

// BenchmarkLRUWith256 runs a benchmark test for the sharded LRU instance
// with 256 sharded [lrucore.LRUCore] instances.
func BenchmarkLRUWith256(b *testing.B) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := tlru.New[int, testutil.User](capacity, tlru.WithShards(256))
		if err != nil {
			b.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}

	keys := 16384
	numOps := 2 << 24

	zipFData := testutil.GenerateZipfData(keys, numOps)
	randomData := testutil.GenerateRandomData(keys, numOps)
	testutil.BenchmarkCache(b, init, "Zipf", 512, numOps, zipFData)
	testutil.BenchmarkCache(b, init, "Random", 512, numOps, randomData)
}

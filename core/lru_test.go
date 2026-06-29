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

// TestRaceLRU runs [raceBase] for different types of keys.
func TestRaceLRU(t *testing.T) {
	raceBase[int32](t, "int32")
	raceBase[int](t, "int")
	raceBase[uint](t, "uint")
	raceBase[string](t, "string")
}

// raceBase runs a concurrency test for the [LRU] instance with keys.
func raceBase[K comparable](t *testing.T, key string) {
	cache, err := core.New[K, testutil.User](2048)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.RunRace(t, key, cache)
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

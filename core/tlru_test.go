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

// TestRaceTLRU runs [raceWithTTL] for different types of keys.
func TestRaceTLRU(t *testing.T) {
	raceWithTTL[int32](t, "int32")
	raceWithTTL[int](t, "int")
	raceWithTTL[uint](t, "uint")
	raceWithTTL[string](t, "string")
}

// raceWithTTL runs a concurrency test for the [TLRU] instance with keys.
func raceWithTTL[K comparable](t *testing.T, key string) {
	cache, err := core.NewWithTTL[K, testutil.User](2048, time.Hour)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.RunRace(t, key, cache)
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

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
		cache, err := core.NewSegmented[int, testutil.User](capacity, 20)
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

// raceSegmented runs a concurrency test for the [SLRU] instance with keys.
func raceSegmented[K comparable](t *testing.T, key string) {
	cache, err := core.NewSegmented[K, testutil.User](2048, 30)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.RunRace(t, key, cache)
}

// BenchmarkSLRU runs a benchmark test for the [SLRU] instance.
func BenchmarkSLRU(b *testing.B) {
	cache, err := core.NewSegmented[int, testutil.User](2048, 20)
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

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru_test

import (
	"testing"
	"time"

	"github.com/justpranavrs/tlru"
	"github.com/justpranavrs/tlru/internal/testutil"
	"github.com/justpranavrs/tlru/mux"
)

// TestPoolTLRU runs a basic unit test for the sharded TLRU instance.
func TestPoolTLRU(t *testing.T) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := tlru.NewWithTTL[int, testutil.User](capacity, time.Hour)
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicLRUData, init)
}

// TestRaceTLRU runs [raceWithTTL] for different types of keys.
func TestRaceTLRU(t *testing.T) {
	raceWithTTL[int32](t, "int32")
	raceWithTTL[int](t, "int")
	raceWithTTL[uint](t, "uint")
	raceWithTTL[string](t, "string")
}

// raceWithTTL runs a concurrency test for the [PoolTLRU] instance with keys.
func raceWithTTL[K comparable](t *testing.T, key string) {
	cache, err := tlru.NewSegmented[K, testutil.User](2048, 30)
	if err != nil {
		t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.RunRace(t, key, cache)
}

// FuzzPoolTLRU runs a fuzz test for the sharded TLRU instance.
func FuzzPoolTLRU(f *testing.F) {
	hasher := mux.NewMH32[int](tlru.DefaultShards)
	cache, err := tlru.NewWithTTL[int, testutil.User](512, time.Hour, tlru.WithMux(hasher))
	if err != nil {
		f.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
	}
	testutil.FuzzCache(f, cache, hasher, 8192, 512, 1536, tlru.DefaultShards)
}

// BenchmarkPoolTLRUWith64 runs a benchmark test for the sharded TLRU instance
// with 64 sharded [core.TLRU] instances.
func BenchmarkPoolTLRUWith64(b *testing.B) {
	cache, err := tlru.NewWithTTL[int, testutil.User](16384, time.Hour, tlru.WithShards(64))
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

// BenchmarkPoolTLRU runs a benchmark test for the sharded TLRU instance.
func BenchmarkPoolTLRU(b *testing.B) {
	cache, err := tlru.NewWithTTL[int, testutil.User](16384, time.Hour)
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

// BenchmarkPoolTLRUWith256 runs a benchmark test for the sharded TLRU instance
// with 256 sharded [core.TLRU] instances.
func BenchmarkPoolTLRUWith256(b *testing.B) {
	cache, err := tlru.NewWithTTL[int, testutil.User](16384, time.Hour, tlru.WithShards(256))
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

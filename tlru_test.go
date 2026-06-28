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

// TestRacePoolTLRU_Int runs a concurrency test for the sharded TLRU instance with int keys.
func TestRacePoolTLRU_Int(t *testing.T) {
	cache, err := tlru.NewWithTTL[int, testutil.User](4096, time.Hour)
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

// TestRacePoolTLRU_Int runs a concurrency test for the sharded TLRU instance with int32 keys.
func TestRacePoolTLRU_Int32(t *testing.T) {
	cacheInt32, err := tlru.NewWithTTL[int32, testutil.User](4096, time.Hour)
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

// TestRacePoolTLRU_Int runs a concurrency test for the sharded TLRU instance with uint keys.
func TestRacePoolTLRU_Uint(t *testing.T) {
	cacheUint, err := tlru.NewWithTTL[uint, testutil.User](4096, time.Hour)
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

// TestRacePoolTLRU_Int runs a concurrency test for the sharded TLRU instance with string keys.
func TestRacePoolTLRU_String(t *testing.T) {
	cacheStr, err := tlru.NewWithTTL[string, testutil.User](4096, time.Hour)
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

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lrucore_test

import (
	"testing"

	"github.com/justpranavrs/tlru/internal/testutil"
	"github.com/justpranavrs/tlru/lrucore"
)

// TestLRUCore runs a basic and advanced unit tests for the core LRU instance.
func TestLRUCore(t *testing.T) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := lrucore.New[int, testutil.User](capacity)
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicTestData, init)
	testutil.TestCache(t, testutil.AdvancedTestData, init)
}

// FuzzLRUCore runs a fuzz test for the core LRU instance.
func FuzzLRUCore(f *testing.F) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := lrucore.New[int, testutil.User](capacity)
		if err != nil {
			f.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.FuzzCache(f, init, 512, 8192, 2)
}

// BenchmarkLRUCore runs a benchmark test for the core LRU instance.
func BenchmarkLRUCore(b *testing.B) {
	var init testutil.TestInit = func(capacity int) testutil.CacheTest[int, testutil.User] {
		cache, err := lrucore.New[int, testutil.User](capacity)
		if err != nil {
			b.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}

	keys := 16384
	benchOpsSize := 2 << 24

	zipFData := testutil.GenerateZipfData(keys, benchOpsSize)
	randomData := testutil.GenerateRandomData(keys, benchOpsSize)
	testutil.BenchmarkCache(b, init, "Zipf", 512, benchOpsSize, zipFData)
	testutil.BenchmarkCache(b, init, "Random", 512, benchOpsSize, randomData)
}

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
	var init testutil.TestInit = func(capacity int) tlru.Cache[int, testutil.User] {
		cache, err := tlru.New[int, testutil.User](capacity)
		if err != nil {
			t.Fatalf("[ERROR] could not initialize Cache instance: %v", err)
		}
		return cache
	}
	testutil.TestCache(t, testutil.BasicTestData, init)
}

// BenchmarkLRU runs a benchmark test for the sharded LRU instance.
func BenchmarkLRU(b *testing.B) {
	var init testutil.TestInit = func(capacity int) tlru.Cache[int, testutil.User] {
		cache, err := tlru.New[int, testutil.User](capacity)
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

// BenchmarkLRUWithUnsafe runs a benchmark test for the sharded LRU instance
// with the Unsafe Option.
func BenchmarkLRUWithUnsafe(b *testing.B) {
	var init testutil.TestInit = func(capacity int) tlru.Cache[int, testutil.User] {
		cache, err := tlru.New(capacity, tlru.WithUnsafe[int, testutil.User]())
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

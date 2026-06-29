// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testutil

import (
	"sync"
	"testing"
)

type raceRunner interface {
	Run(t *testing.T, cache any)
}

type raceTestCase[K comparable] struct {
	getKey func(CacheOp) K
}

func RunRace(t *testing.T, key string, cache any) {
	runner, exists := RaceMap[key]
	if !exists {
		t.Fatalf("[ERROR] unknown race test type: %s", key)
	}

	t.Run(key, func(t *testing.T) {
		runner.Run(t, cache)
	})
}

func (r raceTestCase[K]) Run(t *testing.T, cache any) {
	c, ok := cache.(CacheTest[K, User])
	if !ok {
		t.Fatalf("[ERROR] cache type mismatch: want CacheTest[%T, User], got %T", *new(K), cache)
	}
	testRaceCache(c, 16384, 1<<20, 256, r.getKey)
}

var RaceMap = map[string]raceRunner{
	"int": raceTestCase[int]{
		getKey: func(c CacheOp) int {
			return c.Key
		},
	},
	"int32": raceTestCase[int32]{
		getKey: func(c CacheOp) int32 {
			return int32(c.Key)
		},
	},
	"uint": raceTestCase[uint]{
		getKey: func(c CacheOp) uint {
			return uint(c.Key)
		},
	},
	"string": raceTestCase[string]{
		getKey: func(c CacheOp) string {
			return c.Value.Name
		},
	},
}

// testRaceCache is the main concurrency check test for LRU.
// It checks if it leads to data race with the -race flag.
func testRaceCache[K comparable](cache CacheTest[K, User], keys int, numOps int, numWorkers int, getKey func(CacheOp) K) {
	data := GenerateZipfData(keys, numOps)
	batchSize := (numOps / numWorkers)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for w := range numWorkers {
		go func(workerID int) {
			defer wg.Done()

			st := workerID * batchSize
			en := st + batchSize
			if workerID == numWorkers-1 {
				en = numOps
			}

			for i := st; i < en; i++ {
				key := getKey(data[i])
				if data[i].Method == opGet {
					cache.Get(key)
				} else {
					cache.Put(key, data[i].Value)
				}
			}
		}(w)
	}
	wg.Wait()
}

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testutil

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"sync"
	"testing"

	"github.com/justpranavrs/tlru/internal/mathutil"
	"github.com/justpranavrs/tlru/lrucore"
	"github.com/justpranavrs/tlru/mux"
)

// TestInit takes in capacity as input and returns a Cache instance
type TestInit func(int) CacheTest[int, User]

// TestCache is the main unit tests function.
// ops is an array of all the operations.
func TestCache(t *testing.T, ops []testCacheOp, init TestInit) {
	var cache CacheTest[int, User]
	for i, op := range ops {
		switch op.method {
		case opInit:
			cache = init(op.capacity)
		case opCapacity:
			cap := cache.Capacity()
			if cap != op.expectedNumber {
				t.Fatalf("\n[ERROR]:[CAPACITY] expected: %d, value : %d, tick: %d", op.expectedNumber, cap, i)
			}
		case opFlush:
			cache.Flush()
		case opGet:
			val, ok := cache.Get(op.key)
			if ok != op.expectedBool {
				t.Fatalf("\n[ERROR]:[GET]:[CACHED] expected : %t, tick: %d", ok, i)
			}
			if ok && val != op.expectedValue {
				t.Fatalf("\n[ERROR]:[GET]:[VALUE] expected : %s, value : %s, tick: %d", op.expectedValue.Name, val.Name, i)
			}
		case opPeek:
			val, ok := cache.Peek(op.key)
			if ok != op.expectedBool {
				t.Fatalf("\n[ERROR]:[PEEK]:[CACHED] expected : %t, tick: %d", ok, i)
			}
			if ok && val != op.expectedValue {
				t.Fatalf("\n[ERROR]:[PEEK]:[VALUE] expected : %s, value : %s, tick: %d", op.expectedValue.Name, val.Name, i)
			}
		case opPut:
			cache.Put(op.key, op.value)
		case opSize:
			size := cache.Size()
			if size != op.expectedNumber {
				t.Fatalf("\n[ERROR]:[SIZE] expected : %d, value : %d, tick: %d", op.expectedNumber, size, i)
			}
		case opUpsert:
			state, _ := cache.Upsert(op.key, op.value)
			if op.expectedNumber != int(state) {
				t.Fatalf("\n[ERROR]:[UPSERT] expected : %d, value : %d, tick: %d", op.expectedNumber, state, i)
			}
		default:
			t.Fatal("\n[ERROR]:[INVALID METHOD]")
		}
	}
}

// TestRaceCache is the main concurrency check test for LRU.
// It checks if it leads to data race with the -race flag.
func TestRaceCache[K comparable](
	t *testing.T, cache CacheTest[K, User], keys int,
	numOps int, numWorkers int, getKey func(CacheOp) K,
) {
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

// FuzzCache runs a Fuzz test on the Cache instance.
func FuzzCache(f *testing.F, cache CacheTest[int, User], mux mux.Mux[int], numOps int, capacity int, shards int) {
	actionMask := len(actions) - 1
	keyMask := fuzzKeys - 1
	maxSize := capacity / int(shards)

	// each op is nBytes
	seed := make([]byte, numOps*fuzzBytes) // generate fuzz seed
	for i := range seed {
		seed[i] = byte(rand.IntN(256))
	} // initially generate a fuzz seed for 1024 cache operations

	f.Add(seed)
	f.Fuzz(func(t *testing.T, arg []byte) {
		cache.Flush() // the cache

		state := make([]User, fuzzKeys) // source of truth
		tick := make([]int, fuzzKeys)   // to find evict index
		size := make([]int, shards)     // to track size in o(1)
		totalSize := 0

		for i := range fuzzKeys {
			tick[i] = -1
		}

		ops := make([]int, len(arg)/fuzzBytes)
		for i := range ops {
			op := 0
			for j := range fuzzBytes {
				op = op<<8 + int(arg[i*fuzzBytes+j])
			}
			ops[i] = op
		}

		tk := 1
		for _, op := range ops {
			method := actions[op&actionMask] // since actions and keys are 2^n
			key := op & keyMask
			shard := mux(key)

			name := "tlru_user_" + strconv.Itoa(key)
			user := User{
				Name:  name,
				Email: name + "@gmail.com",
			}
			isCached := (tick[key] != -1)

			switch method {
			case opDelete:
				val, ok := cache.Delete(key)
				if ok != isCached {
					t.Fatalf("\n[ERROR] invalid presence of key in [DELETE], expected: %t, tick: %d", ok, tk)
				}
				if ok {
					if val != state[key] {
						t.Fatalf("\n[ERROR] unexpected value found in [DELETE], tick: %d", tk)
					}
					state[key] = User{}
					tick[key] = -1
					size[shard]--
					totalSize--
				}
			case opGet:
				val, ok := cache.Get(key)
				if ok != isCached {
					t.Fatalf("\n[ERROR] invalid presence of key in [GET], expected: %t, tick: %d", ok, tk)
				}
				if ok {
					tick[key] = tk
					if val != state[key] {
						t.Fatalf("\n[ERROR] unexpected value found in [GET], tick: %d", tk)
					}
				}
			case opPeek:
				val, ok := cache.Peek(key)
				if ok != isCached {
					t.Fatalf("\n[ERROR] invalid presence of key in [GET QUIET], expected: %t, tick: %d", ok, tk)
				}
				if ok && val != state[key] {
					t.Fatalf("\n[ERROR] unexpected value found in [GET QUIET], tick: %d", tk)
				}
			case opPut:
				cache.Put(key, user)
				if !isCached {
					if size[shard] >= maxSize {
						idx := evictKey(tick, shard, mux)
						state[idx] = User{}
						tick[idx] = -1
					} else {
						size[shard]++
						totalSize++
					}
				}

				state[key] = user
				tick[key] = tk
			case opSize:
				sz := cache.Size()
				if sz != totalSize {
					t.Fatalf("\n[ERROR] unexpected size of cache, expected: %d, value: %d, tick: %d", totalSize, sz, tk)
				}
			case opUpsert:
				var st lrucore.UpsertState

				ste, _ := cache.Upsert(key, user)
				if !isCached {
					if size[shard] >= maxSize {
						idx := evictKey(tick, shard, mux)
						state[idx] = User{}
						tick[idx] = -1
						st = lrucore.AddOnEvict
					} else {
						size[shard]++
						totalSize++
						st = lrucore.AddNoEvict
					}
				} else {
					st = lrucore.Replace
				}

				if st != ste {
					t.Fatalf("\n[ERROR] unexpected upsert of cache, expected: %d, value: %d, tick: %d", st, ste, tk)
				}

				state[key] = user
				tick[key] = tk
			default:
				t.Fatal("\n[ERROR] invalid test operation method")
			}
			tk++
		}
	})
}

func BenchmarkCache(
	b *testing.B, cache CacheTest[int, User], prefix string, capacity int,
	numOps int, data []CacheOp,
) {
	numOps = mathutil.NextPowerOf2(uint(numOps))
	numOpsMask := numOps - 1
	var sink User

	b.Run(prefix+"_Puts", func(b *testing.B) {
		b.ReportAllocs()
		cache.Flush()
		b.ResetTimer()

		i := 0
		for b.Loop() {
			cache.Put(data[i].Key, data[i].Value)
			i = (i + 1) & numOpsMask
		}
	})

	b.Run(prefix+"_Gets", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		i := 0
		var user User
		for b.Loop() {
			user, _ = cache.Get(data[i].Key)
			i = (i + 1) & numOpsMask
		}
		sink = user
	})

	b.Run(prefix+"_Mixed", func(b *testing.B) {
		b.ReportAllocs()
		cache.Flush()

		for i := 0; cache.Size() < capacity; i++ {
			cache.Put(data[i].Key, data[i].Value)
		}
		b.ResetTimer()

		hits := 0
		total := 0

		i := 0
		var user User
		for b.Loop() {
			d := data[i]
			if d.Method == opGet {
				val, ok := cache.Get(data[i].Key)
				if ok {
					hits++
					user = val
				}
			} else {
				cache.Put(data[i].Key, data[i].Value)
			}
			total++
			i = (i + 1) & numOpsMask
		}

		sink = user

		ratio := float64(hits) / float64(total)
		fmt.Printf("\nHits : %d, Miss : %d, Ratio: %.4f\n", hits, total-hits, ratio)
	})

	b.Run(prefix+"_Mixed_Parallel", func(b *testing.B) {
		b.ReportAllocs()
		cache.Flush()

		for i := 0; cache.Size() < capacity; i++ {
			cache.Put(data[i].Key, data[i].Value)
		}
		b.ResetTimer()
		var user User

		b.RunParallel(func(p *testing.PB) {
			i := rand.IntN(numOps)
			var userP User
			for p.Next() {
				d := data[i]
				if d.Method == opGet {
					val, ok := cache.Get(data[i].Key)
					if ok {
						userP = val
					}
				} else {
					cache.Put(data[i].Key, data[i].Value)
				}
				i = (i + 1) & numOpsMask
			}
			user = userP

		})
		sink = user
	})
	sink = User{Name: sink.Email, Email: sink.Name}
}

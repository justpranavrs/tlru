// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testutil

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/justpranavrs/tlru/internal/mathutil"
)

// CacheTest is a testing interface for the LRU instances.
// For more details, refer [tlru.Cache]
type CacheTest[K comparable, V any] interface {
	Capacity() int
	Compaction()
	Contains(key K) bool
	Flush()
	Get(key K) (V, bool)
	GetQuiet(key K) (V, bool)
	Put(key K, value V)
	PutGrows(key K, value V) bool
	Size() int
}

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
				t.Fatalf("\n[ERROR] discrepancy in capacity. expected: %d, value : %d, tick: %d", op.expectedNumber, cap, i)
			}
		case opContains:
			if cache.Contains(op.key) != op.expectedBool {
				t.Fatalf("\n[ERROR] invalid presence of key. key: %d, expected: %t, tick: %d", op.key, op.expectedBool, i)
			}
		case opFlush:
			cache.Flush()
		case opGet:
			val, ok := cache.Get(op.key)
			if ok != op.expectedBool {
				t.Fatalf("\n[ERROR] invalid presence of key, expected : %t, tick: %d", ok, i)
			}
			if ok && val != op.expectedValue {
				t.Fatalf("\n[ERROR] unexpected value found, expected : %s, value : %s, tick: %d", op.expectedValue.Name, val.Name, i)
			}
		case opGetQuiet:
			val, ok := cache.GetQuiet(op.key)
			if ok != op.expectedBool {
				t.Fatalf("\n[ERROR] invalid presence of key, expected : %t, tick: %d", ok, i)
			}
			if ok && val != op.expectedValue {
				t.Fatalf("\n[ERROR] unexpected value found, expected : %s, value : %s, tick: %d", op.expectedValue.Name, val.Name, i)
			}
		case opPut:
			cache.Put(op.key, op.value)
		case opSize:
			size := cache.Size()
			if size != op.expectedNumber {
				t.Fatalf("\n[ERROR] unexpected value found, expected : %d, value : %d, tick: %d", op.expectedNumber, size, i)
			}
		default:
			t.Fatal("\n[ERROR] invalid test operation method")
		}
	}
}

// FuzzCache runs a Fuzz test on the Cache instance.
func FuzzCache(f *testing.F, init TestInit, capacity int, fuzzOpsSize int, nBytes int) {
	keys := mathutil.NextPowerOf2(1 << ((nBytes << 3) - 4)) // round keys to the next power of 2
	// 4 is for 16 actions in actions array

	evictKey := func(tick []int) int { // finds the evict index
		evictIdx := 0
		evictTick := 1<<31 - 1
		for idx := range keys {
			// if this is the least tick not equal to -1, no two keys can have same tick
			if tick[idx] != -1 && tick[idx] < evictTick {
				evictIdx = idx
				evictTick = tick[idx]
			}
		}
		return evictIdx
	}

	actionMask := len(actions) - 1
	keyMask := keys - 1

	// each op is nBytes
	seed := make([]byte, fuzzOpsSize*nBytes) // generate fuzz seed
	for i := range seed {
		seed[i] = byte(rand.IntN(256))
	} // initially generate a fuzz seed for 1024 cache operations

	f.Add(seed)
	f.Fuzz(func(t *testing.T, arg []byte) {
		cache := init(capacity)     // the cache
		state := make([]User, keys) // source of truth
		tick := make([]int, keys)   // to find evict index
		size := 0                   // to track size in o(1)
		for i := range keys {
			tick[i] = -1
		}

		ops := make([]int, len(arg)/nBytes)
		for i := range ops {
			op := 0
			for j := range nBytes {
				op = op<<8 + int(arg[i*nBytes+j])
			}
			ops[i] = op
		}

		tk := 1
		for _, op := range ops {
			method := actions[op&actionMask] // since actions and keys are 2^n
			key := op & keyMask

			name := fmt.Sprintf("user_%d", key)
			user := User{
				Name:  name,
				Email: name + "@gmail.com",
			}
			isCached := (tick[key] != -1)

			switch method {
			case opCapacity:
				cap := cache.Capacity()
				if cap != capacity {
					t.Fatalf("\n[ERROR] discrepancy in capacity, tick: %d", tk)
				}
			case opContains:
				if cache.Contains(key) != isCached {
					t.Fatalf("\n[ERROR] invalid presence of key in [CONTAINS], expected: %t, tick: %d", !isCached, tk)
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
			case opGetQuiet:
				val, ok := cache.GetQuiet(key)
				if ok != isCached {
					t.Fatalf("\n[ERROR] invalid presence of key in [GET QUIET], expected: %t, tick: %d", ok, tk)
				}
				if ok && val != state[key] {
					t.Fatalf("\n[ERROR] unexpected value found in [GET QUIET], tick: %d", tk)
				}
			case opPut:
				grow := cache.PutGrows(key, user)
				if grow {
					size++
				} else {
					tick[evictKey(tick)] = -1
				}
				state[key] = user
				tick[key] = tk
			case opSize:
				sz := cache.Size()
				if sz != size {
					t.Fatalf("\n[ERROR] unexpected size of cache, expected: %d, value: %d, tick: %d", size, sz, tk)
				}
			default:
				t.Fatal("\n[ERROR] invalid test operation method")
			}
			tk++
		}
	})
}

func BenchmarkCache(b *testing.B, init TestInit, prefix string, capacity int, benchOpsSize int, data []CacheOp) {
	benchOpsSize = mathutil.NextPowerOf2(benchOpsSize)
	benchOpsSizeMask := benchOpsSize - 1

	cache := init(capacity)
	var sink User

	b.Run(prefix+"_Puts", func(b *testing.B) {
		b.ReportAllocs()
		cache.Flush()
		b.ResetTimer()

		i := 0
		for b.Loop() {
			cache.Put(data[i].key, data[i].value)
			i = (i + 1) & benchOpsSizeMask
		}
	})

	b.Run(prefix+"_Gets", func(b *testing.B) {
		b.ReportAllocs()
		cache.Flush()
		b.ResetTimer()

		i := 0
		var user User
		for b.Loop() {
			user, _ = cache.Get(data[i].key)
			i = (i + 1) & benchOpsSizeMask
		}
		sink = user
	})

	b.Run(prefix+"_Mixed", func(b *testing.B) {
		b.ReportAllocs()
		cache.Flush()

		for i := 0; cache.Size() < capacity; i++ {
			cache.Put(data[i].key, data[i].value)
		}
		b.ResetTimer()

		hits := 0
		total := 0

		i := 0
		var user User
		for b.Loop() {
			d := data[i]
			if d.method == opGet {
				val, ok := cache.Get(data[i].key)
				if ok {
					hits++
					user = val
				}
			} else {
				cache.Put(data[i].key, data[i].value)
			}
			total++
			i = (i + 1) & benchOpsSizeMask
		}

		sink = user

		ratio := float64(hits) / float64(total)
		fmt.Printf("Hits : %d, Miss : %d, Ratio: %.4f\n", hits, total-hits, ratio)
	})

	b.Run(prefix+"_Mixed_Parallel", func(b *testing.B) {
		b.ReportAllocs()
		cache.Flush()

		for i := 0; cache.Size() < capacity; i++ {
			cache.Put(data[i].key, data[i].value)
		}
		b.ResetTimer()
		var user User

		b.RunParallel(func(p *testing.PB) {
			i := rand.IntN(benchOpsSize)
			var userP User
			for p.Next() {
				d := data[i]
				if d.method == opGet {
					val, ok := cache.Get(data[i].key)
					if ok {
						userP = val
					}
				} else {
					cache.Put(data[i].key, data[i].value)
				}
				i = (i + 1) & benchOpsSizeMask
			}
			user = userP

		})
		sink = user
	})
	sink = User{Name: sink.Email, Email: sink.Name}
}

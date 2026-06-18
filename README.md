# tlru

[![GoDoc Reference](https://pkg.go.dev/badge/github.com/justpranavrs/tlru)](https://pkg.go.dev/github.com/justpranavrs/tlru)
[![Go Report Card](https://goreportcard.com/badge/github.com/justpranavrs/tlru)](https://goreportcard.com/report/github.com/justpranavrs/tlru)

`tlru` is a high-performance, array-based `LRU` cache for Go with **zero runtime allocations** and **zero dependencies**. It also supports utilizing multiple independent containers to eliminate lock contention and allow high-concurrency operations without bottlenecks.

#### **NOTE**: The current version has no support for `TTL`. It will be added in the future versions.

#### **Built with Go 1.26. Supports Go 1.22+**
## Table of Contents
- [Introduction](#introduction)
    - [How does lrucore.LRUCore work?](#how-does-lrucorelrucore-work)
	- [What is tlru.LRU?](#what-is-tlrulru)
- [Installation](#installation)
- [Examples](#examples)
    - [Basic LRU Cache](#basic-lru-cache)
    - [Customization](#customization)
- [Benchmarks](#benchmarks)
- [License](#license)

## Introduction
### How does `lrucore.LRUCore` work?
- The `lrucore.LRUCore` uses an array-based doubly linked list with int32 indices. This guarantees zero runtime allocations.
- Each of these instances have a mutex lock to ensure safety in concurrent operations.
- `lrucore.LRUCore` has Go's support for generics.

### What is `tlru.LRU`?
- While `lrucore.LRUCore` is incredibly powerful, struggles under heavy concurrent workloads. That is where `tlru.LRU` shines. It uses a sharded architecture, consisting of many `lrucore.LRUCore` instances. Since each Instance is protected by a mutex lock, `tlru.LRU` doesn't need its own mutex lock.
- It doesn't undergo a global based eviction. It uses a `shard-based local eviction` for its keys. The more the shards, the lesser the chance to evict the global least recently used key. To use the global based approach, use `lrucore.LRUCore`.
- It uses a `custom offset FNV-1a` Hash algorithm which is resistant to `Hash DOS attacks`.
- It has two options: 
  - `WithShards`: It allows the user to customize the number of shards `tlru.LRU` creates.

- Both the instances have an experimental feature called `Compaction`. It allows the cache to fix memory fragmentation by using an expensive O(N) call. It can be used on both the instances anytime manually. It does not happen automatically.

## Installation

```bash
go get -u github.com/justpranavrs/tlru
```

## Examples
It is very easy to setup a basic LRU cache instance.
### Basic LRU Cache
```go
package main

import (
	"fmt"

	"github.com/justpranavrs/tlru"
)

func main() {
	// create a new LRU cache instance.
    // default number of containers is 128.
	cache, err := tlru.New[int, int](1000000)
	if err != nil {
		fmt.Printf("lru cache initialization error: %v", err)
	}
	cache.Put(1, 18)

	val, ok := cache.Get(1)
	if !ok {
		fmt.Println("key not present in cache")
	}
	fmt.Println(val) // 18
}
```

### Customization
To customize the Cache
```go
cache, err := tlru.New(cacheCapacity, tlru.WithShards[int, User](64))
```
#### **Note** : For more examples, refer [here](./lru_example_test.go)

## Benchmarks

#### Environment
- os: archlinux/amd64
- cpu : AMD Ryzen 7 260 w/ Radeon 780M Graphics

```
goos: linux
goarch: amd64
pkg: github.com/justpranavrs/tlru
cpu: AMD Ryzen 7 260 w/ Radeon 780M Graphics

BenchmarkLRU/Zipf_Puts-16                           34412677      35.34 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Zipf_Gets-16                          100000000      11.74 ns/op       0 B/op       0 allocs/op

Hits : 10716111, Miss : 23262154, Ratio: 0.315
BenchmarkLRU/Zipf_Mixed-16                          33978265      35.31 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Zipf_Mixed_Parallel-16                 68256360      17.83 ns/op       0 B/op       0 allocs/op

BenchmarkLRU/Random_Puts-16                         27416224      42.84 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Random_Gets-16                         97843783      12.17 ns/op       0 B/op       0 allocs/op

Hits : 533203, Miss : 33664641, Ratio: 0.0156
BenchmarkLRU/Random_Mixed-16                        34197844      34.43 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Random_Mixed_Parallel-16               91967905      12.59 ns/op       0 B/op       0 allocs/op

BenchmarkLRUWith256/Zipf_Puts-16                    33995874      35.09 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Zipf_Gets-16                    97174560      12.23 ns/op       0 B/op       0 allocs/op

Hits : 11193751, Miss : 24574472, Ratio: 0.3130
BenchmarkLRUWith256/Zipf_Mixed-16                   35768223      35.11 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Zipf_Mixed_Parallel-16          73852075      16.76 ns/op       0 B/op       0 allocs/op

BenchmarkLRUWith256/Random_Puts-16                  25831059      45.55 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Random_Gets-16                  86570054      12.94 ns/op       0 B/op       0 allocs/op

Hits : 520365, Miss : 32758787, Ratio: 0.0156
BenchmarkLRUWith256/Random_Mixed-16                 33279152      34.73 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Random_Mixed_Parallel-16       100000000      10.45 ns/op       0 B/op       0 allocs/op
PASS
ok      github.com/justpranavrs/tlru    81.728s

---

goos: linux
goarch: amd64
pkg: github.com/justpranavrs/tlru/lrucore
cpu: AMD Ryzen 7 260 w/ Radeon 780M Graphics

BenchmarkLRUCore/Zipf_Puts-16                       28044559      43.67 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Zipf_Gets-16                      188881784       6.26 ns/op       0 B/op       0 allocs/op

Hits : 9674353, Miss : 20831821, Ratio: 0.3171
BenchmarkLRUCore/Zipf_Mixed-16                      30506174      39.91 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Zipf_Mixed_Parallel-16             13533098      88.38 ns/op       0 B/op       0 allocs/op

BenchmarkLRUCore/Random_Puts-16                     19492653      61.11 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Random_Gets-16                    195713455       6.33 ns/op       0 B/op       0 allocs/op

Hits : 372214, Miss : 23488905, Ratio: 0.0156
BenchmarkLRUCore/Random_Mixed-16                    23861119      49.85 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Random_Mixed_Parallel-16           13831202      82.86 ns/op       0 B/op       0 allocs/op
PASS
ok      github.com/justpranavrs/tlru/lrucore    40.028s
```

## License
Copyright(c) 2026 Pranav R S

Licensed under [BSD-3-Clause](./LICENSE)
# tlru

[![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![GoDoc Reference](https://pkg.go.dev/badge/github.com/justpranavrs/tlru)](https://pkg.go.dev/github.com/justpranavrs/tlru)
[![Go Report Card](https://goreportcard.com/badge/github.com/justpranavrs/tlru)](https://goreportcard.com/report/github.com/justpranavrs/tlru)
[![CI](https://github.com/justpranavrs/tlru/actions/workflows/test.yml/badge.svg)](https://github.com/justpranavrs/tlru/actions/workflows/test.yml)
[![License](https://img.shields.io/github/license/justpranavrs/tlru?color=56BEB8)](https://img.shields.io/github/license/justpranavrs/tlru?color=56BEB8)
[![GitHub stars](https://img.shields.io/github/stars/justpranavrs/tlru)](https://github.com/justpranavrs/tlru/stargazers)

`tlru` is a high-performance, array-based `LRU` cache for Go with **zero runtime allocations** and **zero dependencies**. It also supports utilizing multiple independent shards to eliminate lock contention and allow high-concurrency operations without bottlenecks. It supports the use of operations for batches and allows a lot of customization, without compromising performance.

#### **NOTE**: The current version has no support for `TTL`. It will be added in the future versions.

#### **Built with Go 1.26. Supports Go 1.24+**

## Table of Contents

- [Introduction](#introduction)
  - [How does lrucore.Core work?](#how-does-lrucorelrucore-work)
  - [What is tlru.LRU?](#what-is-tlrulru)
- [Installation](#installation)
- [Examples](#examples)
  - [Basic LRU Cache](#basic-lru-cache)
  - [Customization](#customization)
- [Benchmarks](#benchmarks)
- [License](#license)

## Introduction

### How does `lrucore.Core` work?

- The `lrucore.Core` uses an array-based doubly linked list with int32 indices. This guarantees zero runtime allocations.
- Each of these instances have a mutex lock to ensure safety in concurrent operations.
- `lrucore.Core` has Go's support for generics.
- It has optimized batch operations like `GetMany` and `PutMany` which reduce the locking contention during high workloads. These operations also support callback functions which execute based on certain hook triggers during the transition in the internal state of the cache.

### What is `tlru.LRU`?

- While `lrucore.Core` is incredibly powerful, struggles under heavy concurrent workloads. That is where `tlru.LRU` shines. It uses a sharded architecture, consisting of many `lrucore.Core` instances. Since each Instance is protected by a mutex lock, `tlru.LRU` doesn't need its own mutex lock.
- It doesn't undergo a global based eviction. It uses a `shard-based local eviction` for its keys. The more the shards, the lesser the chance to evict the global least recently used key. To use the global based approach, use `lrucore.Core`.
- It uses a `mux.Mux` to route the key to one of its shards.
- It has two options:
  - `WithShards`: It allows the user to customize the number of shards `tlru.LRU` creates.
  - `WithMux`: It allows the configuration of `mux.Mux`.

### What is a `mux.Mux`?
- A Mux is a router for the shards which uses a hashing algorithm to distribute the keys evenly across the instances.
- The default hashing algorithms provided in this package are `FNV-1a`, `xxHash32` and Go's `hash/maphash`. The last one has support for `float`, `complex` and `struct` which the `FNV-1a` and `xxHash32` don't provide.
- `WithMux` option allows the configuration by passing a custom hash function of type `mux.Mux` to the `LRU`.

For a detailed walkthrough, refer [here](./LRU.md)

## Installation

```bash
go get -u github.com/justpranavrs/tlru@v0.4.0
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
cache, err := tlru.New[int, User](cacheCapacity, tlru.WithShards(64))
```

#### **Note** : For more examples, refer [here](./lru_example_test.go)

## Benchmarks

#### Environment

- os: archlinux/amd64
- cpu : AMD Ryzen 7 260 w/ Radeon 780M Graphics

#### Competitors
- `Gets - 50%, Puts - 50%`
```
Library          Operations       ns/op      B/op     allocs/op
tlru.LRU		 69,126,327       17.81      0        0
FreeLRU_Sharded  25,522,822       46.11      0        0
Phuslu           60,684,256       19.44      0        0
CCache            4,222,813      305.60     90        3
Theine            8,056,616      143.70     16        0
```

- `Gets - 90%, Puts - 10%`
```
Library          Operations       ns/op      B/op     allocs/op
tlru.LRU     	 61,858,590       17.47      0        0
FreeLRU_Sharded  27,903,866       43.42      0        0
Phuslu           74,030,343       17.16      0        0
CCache            7,093,618      165.90     31        2
Theine           16,673,943       82.08      2        0
```

- `Gets - 10%, Puts - 90%`
```
Library          Operations       ns/op      B/op     allocs/op
TLRU_Sharded     64,692,157       23.82      0        0
FreeLRU_Sharded  24,841,749       48.06      0        0
Phuslu           62,875,437       20.51      0        0
CCache            3,477,776      360.60    149        4
Theine            6,261,631      180.00     24        0
```

#### Performance
- `tlru.LRU` with `64` shards and `mux.NewF32` algorithm:
```
BenchmarkLRUWith64/Zipf_Puts-16                       31532216       37.47 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Zipf_Gets-16                      104158029       11.54 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Zipf_Mixed-16                      32247319       36.78 ns/op       0 B/op       0 allocs/op
Hits : 10177212, Miss : 22070107, Ratio: 0.3156

BenchmarkLRUWith64/Zipf_Mixed_Parallel-16             62996503       19.02 ns/op       0 B/op       0 allocs/op

BenchmarkLRUWith64/Random_Puts-16                     27204386       44.24 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Random_Gets-16                    100000000       11.57 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Random_Mixed-16                    32178608       37.22 ns/op       0 B/op       0 allocs/op
Hits : 503870, Miss : 31674738, Ratio: 0.0157

BenchmarkLRUWith64/Random_Mixed_Parallel-16           77972881       15.49 ns/op       0 B/op       0 allocs/op
```

- `tlru.LRU` with `64` shards and `mux.NewX32` algorithm:
```
BenchmarkLRUWith64/Zipf_Puts-16                       33432204       35.96 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Zipf_Gets-16                      100000000       11.19 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Zipf_Mixed-16                      33517881       36.02 ns/op       0 B/op       0 allocs/op
Hits : 10505944, Miss : 23011937, Ratio: 0.3134

BenchmarkLRUWith64/Zipf_Mixed_Parallel-16             64254285       18.47 ns/op       0 B/op       0 allocs/op

BenchmarkLRUWith64/Random_Puts-16                     27679622       43.94 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Random_Gets-16                    100000000       11.24 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Random_Mixed-16                    31981122       35.51 ns/op       0 B/op       0 allocs/op
Hits : 499739, Miss : 31481383, Ratio: 0.0156

BenchmarkLRUWith64/Random_Mixed_Parallel-16           79188930       15.47 ns/op       0 B/op       0 allocs/op
```

- `tlru.LRU` with `64` shards and `mux.NewMH32` algorithm:
```
BenchmarkLRUWith64/Zipf_Puts-16                       33054550       36.26 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Zipf_Gets-16                      100000000       11.28 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Zipf_Mixed-16                      33174410       35.66 ns/op       0 B/op       0 allocs/op
Hits : 10369406, Miss : 22805004, Ratio: 0.3126

BenchmarkLRUWith64/Zipf_Mixed_Parallel-16             58955106       19.96 ns/op       0 B/op       0 allocs/op

BenchmarkLRUWith64/Random_Puts-16                     27217233       44.25 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Random_Gets-16                    100000000       11.55 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Random_Mixed-16                    32497040       37.01 ns/op       0 B/op       0 allocs/op
Hits : 507588, Miss : 31989452, Ratio: 0.0156

BenchmarkLRUWith64/Random_Mixed_Parallel-16           77845632       15.53 ns/op       0 B/op       0 allocs/op
```

- `tlru.LRU` with `128` shards and `mux.NewX32` algorithm:
```
BenchmarkLRU/Zipf_Puts-16                             33516466       35.88 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Zipf_Gets-16                            100000000       10.74 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Zipf_Mixed-16                            34404482       34.58 ns/op       0 B/op       0 allocs/op
Hits : 10630761, Miss : 23773721, Ratio: 0.3090

BenchmarkLRU/Zipf_Mixed_Parallel-16                   60228430       19.66 ns/op       0 B/op       0 allocs/op

BenchmarkLRU/Random_Puts-16                           28159216       42.65 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Random_Gets-16                          100000000       10.83 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Random_Mixed-16                          34258018       33.99 ns/op       0 B/op       0 allocs/op
Hits : 533000, Miss : 33725018, Ratio: 0.0156

BenchmarkLRU/Random_Mixed_Parallel-16                100000000       11.66 ns/op       0 B/op       0 allocs/op
```

- `tlru.LRU` with `256` shards and `mux.NewX32` algorithm:
```
BenchmarkLRUWith256/Zipf_Puts-16                      34538334       34.77 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Zipf_Gets-16                     100000000       10.95 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Zipf_Mixed-16                     35645792       33.64 ns/op       0 B/op       0 allocs/op
Hits : 10676624, Miss : 24969168, Ratio: 0.2995

BenchmarkLRUWith256/Zipf_Mixed_Parallel-16            75248469       16.22 ns/op       0 B/op       0 allocs/op

BenchmarkLRUWith256/Random_Puts-16                    28202634       42.61 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Random_Gets-16                   100000000       11.24 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Random_Mixed-16                   35532368       33.46 ns/op       0 B/op       0 allocs/op
Hits : 557243, Miss : 34975125, Ratio: 0.0157

BenchmarkLRUWith256/Random_Mixed_Parallel-16         121141429       9.901 ns/op       0 B/op       0 allocs/op
```

- `lrucore.Core`:
```
BenchmarkLRUCore/Zipf_Puts-16                         28712246       41.79 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Zipf_Gets-16                        198613611       6.069 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Zipf_Mixed-16                        30072946       40.02 ns/op       0 B/op       0 allocs/op
Hits : 9533449, Miss : 20539497, Ratio: 0.3170

BenchmarkLRUCore/Zipf_Mixed_Parallel-16               11084142       91.50 ns/op       0 B/op       0 allocs/op

BenchmarkLRUCore/Random_Puts-16                       21257362       56.26 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Random_Gets-16                      193248979       6.135 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Random_Mixed-16                      25733350       45.89 ns/op       0 B/op       0 allocs/op
Hits : 402293, Miss : 25331057, Ratio: 0.0156

BenchmarkLRUCore/Random_Mixed_Parallel-16             13872712       85.06 ns/op       0 B/op       0 allocs/op
```

## License

Copyright(c) 2026 [Pranav R S](https://github.com/justpranavrs)

Licensed under [BSD-3-Clause](./LICENSE)

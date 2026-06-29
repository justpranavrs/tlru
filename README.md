# tlru

[![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![GoDoc Reference](https://pkg.go.dev/badge/github.com/justpranavrs/tlru)](https://pkg.go.dev/github.com/justpranavrs/tlru)
[![Go Report Card](https://goreportcard.com/badge/github.com/justpranavrs/tlru)](https://goreportcard.com/report/github.com/justpranavrs/tlru)
[![CI](https://github.com/justpranavrs/tlru/actions/workflows/test.yml/badge.svg)](https://github.com/justpranavrs/tlru/actions/workflows/test.yml)
[![License](https://img.shields.io/github/license/justpranavrs/tlru?color=56BEB8)](https://img.shields.io/github/license/justpranavrs/tlru?color=56BEB8)
[![GitHub stars](https://img.shields.io/github/stars/justpranavrs/tlru)](https://github.com/justpranavrs/tlru/stargazers)

`tlru` is a high-performance, array-based, time-aware `LRU` cache for Go with **zero runtime allocations** and **zero dependencies**. It also supports utilizing multiple independent shards to eliminate lock contention and allow high-concurrency operations without bottlenecks. It supports the use of operations for batches and allows a lot of customization, without compromising performance.

`tlru` supports `four` main caches:
* Least Recently Used (LRU).
* Time-Aware (LRU).
* Segmented (LRU).
* Time-Aware Segmented (LRU).

`tlru`'s features:
* Simple to Use API.
* Fast.
* Supports Generics.
* Concurrent, Thread-Safe.
* GC-Friendly / Zero Allocations.
* Zero Dependencies.
* Customization of the Sharding Hash Algorithm.
* Batch Operations` for `tlru/core.
* Has Both Absolute and Sliding TTL.
* Customization of Background Clock for TTL.

`tlru`'s limitations:
* Fixed Static Capacity
  * Once the Cache has been initialized, It cannot be resized.
* Lazy Eviction
  * `TTL` evicts the key only if a `Get` or `Peek` operation is executed. It does not automatically free the memory using a background routine.

#### **Built with Go 1.24. Supports Go 1.24+**

## Table of Contents

* [Introduction](#introduction)
  * [How does core.LRU work?](#how-does-lrucorecore-work)
  * [What is tlru.PoolLRU?](#what-is-tlrulru)
  * [What is a mux.Mux?](#what-is-a-muxmux)
* [Installation](#installation)
* [Examples](#examples)
  * [Basic LRU Cache](#basic-lru-cache)
  * [Customization](#customization)
* [Benchmarks](#benchmarks)
* [License](#license)

## Introduction

### How does `core.LRU` work?

* The `core.LRU` uses an array-based doubly linked list with int32 indices. This guarantees zero runtime allocations.
* Each of these instances have a mutex lock to ensure safety in concurrent operations.
* `core.LRU` has Go's support for generics.
* It has optimized batch operations like `GetMany` and `PutMany` which reduce the locking contention during high workloads. This is only limited to `core.LRU`.

### What is `tlru.PoolLRU`?

* While `core.LRU` is incredibly powerful, it struggles under heavy concurrent workloads. That is where `tlru.PoolLRU` shines. It uses a sharded architecture, consisting of many `core.LRU` instances. Since each Instance is protected by a mutex lock, `tlru.PoolLRU` doesn't need its own mutex lock.
* It doesn't undergo a global based eviction. It uses a `shard-based local eviction` for its keys. The more the shards, the lesser the chance to evict the global least recently used key. To use the global based approach, use `core.LRU`.
* It uses a `mux.Mux` to route the key to one of its shards.
* It has two options:
  * `WithShards`: It allows the user to customize the number of shards `tlru.PoolLRU` creates.
  * `WithMux`: It allows the configuration of `mux.Mux`.

### What is a `mux.Mux`?
* A Mux is a router for the shards which uses a hashing algorithm to distribute the keys evenly across the instances.
* The default hashing algorithms provided in this package are `FNV-1a`, `xxHash32` and Go's `hash/maphash`. The last one has support for `float`, `complex` and `struct` which the `FNV-1a` and `xxHash32` don't provide.
* `WithMux` option allows the configuration by passing a custom hash function of type `mux.Mux` to the `LRU`.

### How does `TTL` work?
* `tlru.PoolTLRU` and `core.TLRU` are TTL implementations of `tlru.PoolLRU` and `core.LRU` respectively. They use `Absolute TTL`. The timestamp of a `key` in the cache is updated only on `Put` operations and never on `Get` operations.
* `WithSliding` on these instances enable `Sliding TTL` instead of the default `Absolute TTL`. `Sliding TTL` ensures timestamp updates on `Get` and `Peek` operations too.
* `TTL` works on Lazy Eviction. It is only evicted when a `Get` or `Peek` operation is made to it.

### What is `SLRU`?
* `tlru.PoolSLRU` and `core.SLRU` are implementations of `Segmented` Least Recently Used Cache. They use Two LRUs, Probationary and Protected to avoid `Sequential Scan` Pollution. They offer better hit rates than `LRU` with almost the same speeds. Check the [benchmarks](#benchmarks) below for comparisons.
* SLRU also has `tlru.PoolTSLRU` and `core.TSLRU` which supports `TTL` for `SLRU`.
* Currently it only supports `core.PromotionGet`. It produces better hit rates on tests but it would not be the best option on all scenarios. The support for `core.PromotionGetAndPut` will soon be extended.

For a detailed walkthrough, refer [here](./LRU.md)

## Installation

```bash
go get -u github.com/justpranavrs/tlru@v0.8.0
```

## Examples

It is pretty simple to initialize a TLRU instance.
### Create TLRU Cache

```go
package main

import (
	"fmt"
  "time"

	"github.com/justpranavrs/tlru"
)

func main() {
	// create a new TLRU cache instance.
  // default number of containers is 128.
	cache, err := tlru.NewWithTTL[int, int](51200, 24 * time.Hour)
	if err != nil {
		fmt.Printf("tlru cache initialization error: %v", err)
	}
	cache.Put(1, 18)

	val, ok := cache.Get(1)
	if !ok {
		fmt.Println("key not present in cache")
	}
	fmt.Println(val) // 18
}
```

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
cache, err := tlru.New[int, string](25600, tlru.WithShards(64))
```

#### **Note** : For more examples, refer [here](./lru_example_test.go)

## Benchmarks

### Environment

* os: archlinux/amd64
* cpu: AMD Ryzen 7 260 w/ Radeon 780M Graphics

### Performance

* `tlru.PoolLRU` with `64` shards:
```text
[ Zipf Data ]
  Puts-16                             16104564       72.41 ns/op       0 B/op       0 allocs/op
  Gets-16                             26864718       46.20 ns/op       0 B/op       0 allocs/op
  Mixed-16                            20306064       66.04 ns/op       0 B/op       0 allocs/op
    Hits : 7634801, Miss : 12671263, Ratio: 0.3760
  Mixed_Parallel-16                   46368037       23.71 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                              8980935      113.0 ns/op       0 B/op       0 allocs/op
  Gets-16                             23747763       49.17 ns/op       0 B/op       0 allocs/op
  Mixed-16                            15812216       80.96 ns/op       0 B/op       0 allocs/op
    Hits : 254539, Miss : 15557677, Ratio: 0.0161
  Mixed_Parallel-16                   50857604       22.21 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolLRU` with `128` shards (Default):
```text
[ Zipf Data ]
  Puts-16                             17643631       68.20 ns/op       0 B/op       0 allocs/op
  Gets-16                             24197906       46.90 ns/op       0 B/op       0 allocs/op
  Mixed-16                            19800127       59.71 ns/op       0 B/op       0 allocs/op
    Hits : 7439604, Miss : 12360523, Ratio: 0.3757
  Mixed_Parallel-16                   63165710       19.11 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             12773164       94.57 ns/op       0 B/op       0 allocs/op
  Gets-16                             30141156       42.82 ns/op       0 B/op       0 allocs/op
  Mixed-16                            14883860       83.45 ns/op       0 B/op       0 allocs/op
    Hits : 240647, Miss : 14643213, Ratio: 0.0162
  Mixed_Parallel-16                   69202970       16.92 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolLRU` with `256` shards:
```text
[ Zipf Data ]
  Puts-16                             20157860       63.69 ns/op       0 B/op       0 allocs/op
  Gets-16                             24819087       50.95 ns/op       0 B/op       0 allocs/op
  Mixed-16                            21572820       76.77 ns/op       0 B/op       0 allocs/op
    Hits : 8103629, Miss : 13469191, Ratio: 0.3756
  Mixed_Parallel-16                   71145884       17.56 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                              8973100      150.4 ns/op       0 B/op       0 allocs/op
  Gets-16                             23284514       48.51 ns/op       0 B/op       0 allocs/op
  Mixed-16                            11501730      107.5 ns/op       0 B/op       0 allocs/op
    Hits : 189040, Miss : 11312690, Ratio: 0.0164
  Mixed_Parallel-16                   59917929       22.25 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolSLRU` with `64` shards: (20% Probationary, 80% Protected)
```text
[ Zipf Data ]
  Puts-16                             18153914       64.83 ns/op       0 B/op       0 allocs/op
  Gets-16                             27780100       42.04 ns/op       0 B/op       0 allocs/op
  Mixed-16                            17329540       79.76 ns/op       0 B/op       0 allocs/op
    Hits : 7175125, Miss : 10154415, Ratio: 0.4140
  Mixed_Parallel-16                   55270350       23.02 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             14949028       79.46 ns/op       0 B/op       0 allocs/op
  Gets-16                             22615257       46.90 ns/op       0 B/op       0 allocs/op
  Mixed-16                            16068453       75.12 ns/op       0 B/op       0 allocs/op
    Hits : 399880, Miss : 15668573, Ratio: 0.0249
  Mixed_Parallel-16                   59510073       22.85 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolSLRU` with `128` shards (Default): (20% Probationary, 80% Protected)
```text
[ Zipf Data ]
  Puts-16                             15895562       76.13 ns/op       0 B/op       0 allocs/op
  Gets-16                             26625468       43.91 ns/op       0 B/op       0 allocs/op
  Mixed-16                            19265151       69.79 ns/op       0 B/op       0 allocs/op
    Hits : 7951649, Miss : 11313502, Ratio: 0.4127
  Mixed_Parallel-16                   59550160       18.72 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             15921451       74.82 ns/op       0 B/op       0 allocs/op
  Gets-16                             20015868       50.14 ns/op       0 B/op       0 allocs/op
  Mixed-16                            13137946       94.44 ns/op       0 B/op       0 allocs/op
    Hits : 473612, Miss : 12664334, Ratio: 0.0360
  Mixed_Parallel-16                   72539188       17.57 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolSLRU` with `256` shards: (20% Probationary, 80% Protected)
```text
[ Zipf Data ]
  Puts-16                             19895728       59.64 ns/op       0 B/op       0 allocs/op
  Gets-16                             26094400       40.30 ns/op       0 B/op       0 allocs/op
  Mixed-16                            17964036       85.22 ns/op       0 B/op       0 allocs/op
    Hits : 7398623, Miss : 10565413, Ratio: 0.4119
  Mixed_Parallel-16                   65805986       18.76 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             13133607       86.89 ns/op       0 B/op       0 allocs/op
  Gets-16                             19804249       60.72 ns/op       0 B/op       0 allocs/op
  Mixed-16                             7537466      146.9 ns/op       0 B/op       0 allocs/op
    Hits : 248915, Miss : 7288551, Ratio: 0.0330
  Mixed_Parallel-16                   80201294       15.04 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolTLRU` with `64` shards:
```text
[ Zipf Data ]
  Puts-16                             13442264       91.86 ns/op       0 B/op       0 allocs/op
  Gets-16                             19256506       68.37 ns/op       0 B/op       0 allocs/op
  Mixed-16                            13376530      101.0 ns/op       0 B/op       0 allocs/op
    Hits : 5029620, Miss : 8346910, Ratio: 0.3760
  Mixed_Parallel-16                   38738343       32.15 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                              9877905      122.0 ns/op       0 B/op       0 allocs/op
  Gets-16                             27354732       43.71 ns/op       0 B/op       0 allocs/op
  Mixed-16                            13270051       93.33 ns/op       0 B/op       0 allocs/op
    Hits : 214291, Miss : 13055760, Ratio: 0.0161
  Mixed_Parallel-16                   52081365       20.29 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolTLRU` with `128` shards (Default):
```text
[ Zipf Data ]
  Puts-16                             11024048      112.2 ns/op       0 B/op       0 allocs/op
  Gets-16                             18685392       73.58 ns/op       0 B/op       0 allocs/op
  Mixed-16                            14125198       99.35 ns/op       0 B/op       0 allocs/op
    Hits : 5309647, Miss : 8815551, Ratio: 0.3759
  Mixed_Parallel-16                   38309698       30.31 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                              9858103      161.3 ns/op       0 B/op       0 allocs/op
  Gets-16                             21844707       53.23 ns/op       0 B/op       0 allocs/op
  Mixed-16                            13403178       89.79 ns/op       0 B/op       0 allocs/op
    Hits : 218012, Miss : 13185166, Ratio: 0.0163
  Mixed_Parallel-16                   60236950       19.67 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolTLRU` with `256` shards:
```text
[ Zipf Data ]
  Puts-16                             13183874      111.3 ns/op       0 B/op       0 allocs/op
  Gets-16                             18247238       82.00 ns/op       0 B/op       0 allocs/op
  Mixed-16                            10512177      106.5 ns/op       0 B/op       0 allocs/op
    Hits : 3949932, Miss : 6562245, Ratio: 0.3757
  Mixed_Parallel-16                   45411861       27.15 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                              9252912      135.2 ns/op       0 B/op       0 allocs/op
  Gets-16                             24938228       45.82 ns/op       0 B/op       0 allocs/op
  Mixed-16                            11032528      102.8 ns/op       0 B/op       0 allocs/op
    Hits : 180789, Miss : 10851739, Ratio: 0.0164
  Mixed_Parallel-16                   68274302       20.42 ns/op       0 B/op       0 allocs/op
```

* `core.LRU` (Single-Threaded):
```text
[ Zipf Data ]
  Puts-16                             20896209       50.83 ns/op       0 B/op       0 allocs/op
  Gets-16                             36120510       31.27 ns/op       0 B/op       0 allocs/op
  Mixed-16                            25379311       44.05 ns/op       0 B/op       0 allocs/op
    Hits : 8749268, Miss : 16630043, Ratio: 0.3447
  Mixed_Parallel-16                   13617330       94.98 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             15925435       73.84 ns/op       0 B/op       0 allocs/op
  Gets-16                             40320181       29.28 ns/op       0 B/op       0 allocs/op
  Mixed-16                            19597305       57.60 ns/op       0 B/op       0 allocs/op
    Hits : 309061, Miss : 19288244, Ratio: 0.0158
  Mixed_Parallel-16                   11836870      101.3 ns/op       0 B/op       0 allocs/op
```

* `core.SLRU` (Single-Threaded): (20% Probationary, 80% Protected)
```text
[ Zipf Data ]
  Puts-16                             23328810       51.77 ns/op       0 B/op       0 allocs/op
  Gets-16                             38337645       26.96 ns/op       0 B/op       0 allocs/op
  Mixed-16                            18752710       61.44 ns/op       0 B/op       0 allocs/op
    Hits : 7806166, Miss : 10946544, Ratio: 0.4163
  Mixed_Parallel-16                   13210502      108.3 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             14144305       78.49 ns/op       0 B/op       0 allocs/op
  Gets-16                             47187762       25.92 ns/op       0 B/op       0 allocs/op
  Mixed-16                            24323642       58.47 ns/op       0 B/op       0 allocs/op
    Hits : 82693, Miss : 24240949, Ratio: 0.0034
  Mixed_Parallel-16                   12072326      126.8 ns/op       0 B/op       0 allocs/op
```

* `core.TLRU` (Single-Threaded):
```text
[ Zipf Data ]
  Puts-16                             13879933       85.95 ns/op       0 B/op       0 allocs/op
  Gets-16                             24454734       48.85 ns/op       0 B/op       0 allocs/op
  Mixed-16                            16508620       72.12 ns/op       0 B/op       0 allocs/op
    Hits : 5696291, Miss : 10812329, Ratio: 0.3450
  Mixed_Parallel-16                   10024528      117.6 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             10784348      112.2 ns/op       0 B/op       0 allocs/op
  Gets-16                             39493549       30.14 ns/op       0 B/op       0 allocs/op
  Mixed-16                            15926698       74.94 ns/op       0 B/op       0 allocs/op
    Hits : 249577, Miss : 15677121, Ratio: 0.0157
  Mixed_Parallel-16                   11060875      113.5 ns/op       0 B/op       0 allocs/op
```

## License

Copyright(c) 2026 [Pranav R S](https://github.com/justpranavrs)

Licensed under [BSD-3-Clause](./LICENSE)

# tlru

[![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![GoDoc Reference](https://pkg.go.dev/badge/github.com/justpranavrs/tlru)](https://pkg.go.dev/github.com/justpranavrs/tlru)
[![Go Report Card](https://goreportcard.com/badge/github.com/justpranavrs/tlru)](https://goreportcard.com/report/github.com/justpranavrs/tlru)
[![CI](https://github.com/justpranavrs/tlru/actions/workflows/test.yml/badge.svg)](https://github.com/justpranavrs/tlru/actions/workflows/test.yml)
[![License](https://img.shields.io/github/license/justpranavrs/tlru?color=56BEB8)](https://img.shields.io/github/license/justpranavrs/tlru?color=56BEB8)
[![GitHub stars](https://img.shields.io/github/stars/justpranavrs/tlru)](https://github.com/justpranavrs/tlru/stargazers)

`tlru` is a high-performance, array-based, time-aware `LRU` cache for Go with **zero runtime allocations** and **zero dependencies**. It also supports utilizing multiple independent shards to eliminate lock contention and allow high-concurrency operations without bottlenecks. It supports the use of operations for batches and allows a lot of customization, without compromising performance.

#### **Built with Go 1.24. Supports Go 1.24+**

## Table of Contents

- [Introduction](#introduction)
  - [How does core.LRU work?](#how-does-lrucorecore-work)
  - [What is tlru.PoolLRU?](#what-is-tlrulru)
  - [What is a mux.Mux?](#what-is-a-muxmux)
- [Installation](#installation)
- [Examples](#examples)
  - [Basic LRU Cache](#basic-lru-cache)
  - [Customization](#customization)
- [Benchmarks](#benchmarks)
- [License](#license)

## Introduction

### How does `core.LRU` work?

- The `core.LRU` uses an array-based doubly linked list with int32 indices. This guarantees zero runtime allocations.
- Each of these instances have a mutex lock to ensure safety in concurrent operations.
- `core.LRU` has Go's support for generics.
- It has optimized batch operations like `GetMany` and `PutMany` which reduce the locking contention during high workloads. This is only limited to `core.LRU`.

### What is `tlru.PoolLRU`?

- While `core.LRU` is incredibly powerful, it struggles under heavy concurrent workloads. That is where `tlru.PoolLRU` shines. It uses a sharded architecture, consisting of many `core.LRU` instances. Since each Instance is protected by a mutex lock, `tlru.PoolLRU` doesn't need its own mutex lock.
- It doesn't undergo a global based eviction. It uses a `shard-based local eviction` for its keys. The more the shards, the lesser the chance to evict the global least recently used key. To use the global based approach, use `core.LRU`.
- It uses a `mux.Mux` to route the key to one of its shards.
- It has two options:
  - `WithShards`: It allows the user to customize the number of shards `tlru.PoolLRU` creates.
  - `WithMux`: It allows the configuration of `mux.Mux`.

### What is a `mux.Mux`?
- A Mux is a router for the shards which uses a hashing algorithm to distribute the keys evenly across the instances.
- The default hashing algorithms provided in this package are `FNV-1a`, `xxHash32` and Go's `hash/maphash`. The last one has support for `float`, `complex` and `struct` which the `FNV-1a` and `xxHash32` don't provide.
- `WithMux` option allows the configuration by passing a custom hash function of type `mux.Mux` to the `LRU`.

### How does `TTL` work?
- `tlru.PoolTLRU` and `core.TLRU` are TTL implementations of `tlru.PoolLRU` and `core.LRU` respectively. They use `Absolute TTL`. The timestamp of a `key` in the cache is updated only on `Put` operations and never on `Get` operations.
- `WithSliding` on these instances enable `Sliding TTL` instead of the default `Absolute TTL`. `Sliding TTL` ensures timestamp updates on `Get` and `Peek` operations too.

For a detailed walkthrough, refer [here](./LRU.md)

### What is `SLRU`?
- `tlru.PoolSLRU` and `core.SLRU` are implementations of `Segmented` Least Recently Used Cache. They use Two LRUs, Probationary and Protected to avoid `Sequential Scan` Pollution. They offer better hit rates than `LRU` with almost the same speeds. Check the [benchmarks](#benchmarks) below for comparisons.
- SLRU also has `tlru.PoolTSLRU` and `core.TSLRU` which supports `TTL` for `SLRU`.

## Installation

```bash
go get -u github.com/justpranavrs/tlru@v0.6.1
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

- os: archlinux/amd64
- cpu: AMD Ryzen 7 260 w/ Radeon 780M Graphics

### Performance

* `tlru.PoolLRU` with `64` shards:
```text
[ Zipf Data ]
  Puts-16                             20448162       57.85 ns/op       0 B/op       0 allocs/op
  Gets-16                             28782603       42.00 ns/op       0 B/op       0 allocs/op
  Mixed-16                            24060603       50.85 ns/op       0 B/op       0 allocs/op
    Hits : 9048883, Miss : 15011720, Ratio: 0.3761
  Mixed_Parallel-16                   56745294       20.11 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             14465797       85.14 ns/op       0 B/op       0 allocs/op
  Gets-16                             27400269       40.10 ns/op       0 B/op       0 allocs/op
  Mixed-16                            17655925       68.27 ns/op       0 B/op       0 allocs/op
    Hits : 283844, Miss : 17372081, Ratio: 0.0161
  Mixed_Parallel-16                   55811613       21.13 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolLRU` with `128` shards (Default):
```text
[ Zipf Data ]
  Puts-16                             19976899       58.14 ns/op       0 B/op       0 allocs/op
  Gets-16                             27595329       43.37 ns/op       0 B/op       0 allocs/op
  Mixed-16                            23766600       52.33 ns/op       0 B/op       0 allocs/op
    Hits : 8935425, Miss : 14831175, Ratio: 0.3760
  Mixed_Parallel-16                   55729866       18.62 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             14392702       86.28 ns/op       0 B/op       0 allocs/op
  Gets-16                             28541719       41.17 ns/op       0 B/op       0 allocs/op
  Mixed-16                            17429784       69.39 ns/op       0 B/op       0 allocs/op
    Hits : 278146, Miss : 17151638, Ratio: 0.0160
  Mixed_Parallel-16                   64613409       17.03 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolLRU` with `256` shards:
```text
[ Zipf Data ]
  Puts-16                             19067328       59.72 ns/op       0 B/op       0 allocs/op
  Gets-16                             27279242       44.10 ns/op       0 B/op       0 allocs/op
  Mixed-16                            23164617       54.09 ns/op       0 B/op       0 allocs/op
    Hits : 8706028, Miss : 14458589, Ratio: 0.3758
  Mixed_Parallel-16                   53862920       20.44 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             14246808       87.41 ns/op       0 B/op       0 allocs/op
  Gets-16                             27631425       41.48 ns/op       0 B/op       0 allocs/op
  Mixed-16                            17302650       69.64 ns/op       0 B/op       0 allocs/op
    Hits : 278030, Miss : 17024620, Ratio: 0.0161
  Mixed_Parallel-16                   77106094       15.68 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolSLRU` with `64` shards:
```text
[ Zipf Data ]
  Puts-16                             20233304       59.34 ns/op       0 B/op       0 allocs/op
  Gets-16                             28390084       38.79 ns/op       0 B/op       0 allocs/op
  Mixed-16                            20623526       62.75 ns/op       0 B/op       0 allocs/op
    Hits : 8564859, Miss : 12058667, Ratio: 0.4153
  Mixed_Parallel-16                   43038800       26.30 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             16329735       73.68 ns/op       0 B/op       0 allocs/op
  Gets-16                             24206350       49.58 ns/op       0 B/op       0 allocs/op
  Mixed-16                            17003930       75.44 ns/op       0 B/op       0 allocs/op
    Hits : 659761, Miss : 16344169, Ratio: 0.0388
  Mixed_Parallel-16                   58764945       18.99 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolSLRU` with `128` shards (Default):
```text
[ Zipf Data ]
  Puts-16                             19188867       61.47 ns/op       0 B/op       0 allocs/op
  Gets-16                             27964281       41.10 ns/op       0 B/op       0 allocs/op
  Mixed-16                            20118886       63.12 ns/op       0 B/op       0 allocs/op
    Hits : 8324238, Miss : 11794648, Ratio: 0.4138
  Mixed_Parallel-16                   59818200       19.34 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             17174388       71.47 ns/op       0 B/op       0 allocs/op
  Gets-16                             24535849       48.56 ns/op       0 B/op       0 allocs/op
  Mixed-16                            17497125       74.49 ns/op       0 B/op       0 allocs/op
    Hits : 686760, Miss : 16810365, Ratio: 0.0392
  Mixed_Parallel-16                   73351990       16.57 ns/op       0 B/op       0 allocs/op
```

* `tlru.PoolSLRU` with `256` shards:
```text
[ Zipf Data ]
  Puts-16                             19855791       59.43 ns/op       0 B/op       0 allocs/op
  Gets-16                             26206083       39.29 ns/op       0 B/op       0 allocs/op
  Mixed-16                            19015302       66.13 ns/op       0 B/op       0 allocs/op
    Hits : 7858627, Miss : 11156675, Ratio: 0.4133
  Mixed_Parallel-16                   60483258       18.26 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             16031120       72.16 ns/op       0 B/op       0 allocs/op
  Gets-16                             23839788       49.89 ns/op       0 B/op       0 allocs/op
  Mixed-16                            16771458       74.78 ns/op       0 B/op       0 allocs/op
    Hits : 661508, Miss : 16109950, Ratio: 0.0394
  Mixed_Parallel-16                   83633029       14.34 ns/op       0 B/op       0 allocs/op
```

* `core.LRU` (Single-Threaded):
```text
[ Zipf Data ]
  Puts-16                             24887101       46.92 ns/op       0 B/op       0 allocs/op
  Gets-16                             39485794       31.18 ns/op       0 B/op       0 allocs/op
  Mixed-16                            30837963       39.13 ns/op       0 B/op       0 allocs/op
    Hits : 10636535, Miss : 20201428, Ratio: 0.3449
  Mixed_Parallel-16                   12050385       99.72 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             15438020       75.99 ns/op       0 B/op       0 allocs/op
  Gets-16                             42075907       25.97 ns/op       0 B/op       0 allocs/op
  Mixed-16                            20573970       58.11 ns/op       0 B/op       0 allocs/op
    Hits : 323131, Miss : 20250839, Ratio: 0.0157
  Mixed_Parallel-16                   10230447      116.80 ns/op       0 B/op       0 allocs/op
```

* `core.SLRU` (Single-Threaded):
```text
[ Zipf Data ]
  Puts-16                             24416108       48.75 ns/op       0 B/op       0 allocs/op
  Gets-16                             42146442       25.66 ns/op       0 B/op       0 allocs/op
  Mixed-16                            25450512       50.41 ns/op       0 B/op       0 allocs/op
    Hits : 10583646, Miss : 14866866, Ratio: 0.4159
  Mixed_Parallel-16                   11781547       96.35 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             22181022       53.05 ns/op       0 B/op       0 allocs/op
  Gets-16                             44981145       23.45 ns/op       0 B/op       0 allocs/op
  Mixed-16                            25975788       46.36 ns/op       0 B/op       0 allocs/op
    Hits : 88415, Miss : 25887373, Ratio: 0.0034
  Mixed_Parallel-16                   11981293      109.30 ns/op       0 B/op       0 allocs/op
```

## License

Copyright(c) 2026 [Pranav R S](https://github.com/justpranavrs)

Licensed under [BSD-3-Clause](./LICENSE)

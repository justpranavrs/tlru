# tlru

[![Made with Go](https://img.shields.io/badge/Made%20with-Go-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![GoDoc Reference](https://pkg.go.dev/badge/github.com/justpranavrs/tlru)](https://pkg.go.dev/github.com/justpranavrs/tlru)
[![Go Report Card](https://goreportcard.com/badge/github.com/justpranavrs/tlru)](https://goreportcard.com/report/github.com/justpranavrs/tlru)
[![CI](https://github.com/justpranavrs/tlru/actions/workflows/test.yml/badge.svg)](https://github.com/justpranavrs/tlru/actions/workflows/test.yml)
[![License](https://img.shields.io/github/license/justpranavrs/tlru?color=56BEB8)](https://img.shields.io/github/license/justpranavrs/tlru?color=56BEB8)
[![GitHub stars](https://img.shields.io/github/stars/justpranavrs/tlru)](https://github.com/justpranavrs/tlru/stargazers)

`tlru` is a high-performance, array-based `LRU` cache for Go with **zero runtime allocations** and **zero dependencies**. It also supports utilizing multiple independent shards to eliminate lock contention and allow high-concurrency operations without bottlenecks. It supports the use of operations for batches and allows a lot of customization, without compromising performance.

#### **NOTE**: The current version has no support for `TTL`. It will be added in the future versions.

#### **Built with Go 1.24. Supports Go 1.24+**

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
- It has optimized batch operations like `GetMany` and `PutMany` which reduce the locking contention during high workloads.

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

### Environment

- os: archlinux/amd64
- cpu : AMD Ryzen 7 260 w/ Radeon 780M Graphics

### Performance
* `tlru.LRU` with `64` shards and `mux.NewX32` algorithm:
```text
[ Zipf Workload ]
  Puts-16                             33432204       35.96 ns/op       0 B/op       0 allocs/op
  Gets-16                            100000000       11.19 ns/op       0 B/op       0 allocs/op
  Mixed-16                            33517881       36.02 ns/op       0 B/op       0 allocs/op
    Hits : 10505944, Miss : 23011937, Ratio: 0.3134
  Mixed_Parallel-16                   64254285       18.47 ns/op       0 B/op       0 allocs/op

[ Random Workload ]
  Puts-16                             27679622       43.94 ns/op       0 B/op       0 allocs/op
  Gets-16                            100000000       11.24 ns/op       0 B/op       0 allocs/op
  Mixed-16                            31981122       35.51 ns/op       0 B/op       0 allocs/op
    Hits : 499739, Miss : 31481383, Ratio: 0.0156
  Mixed_Parallel-16                   79188930       15.47 ns/op       0 B/op       0 allocs/op

```

* `tlru.LRU` with `64` shards and `mux.NewMH32` algorithm:

```text
[ Zipf Workload ]
  Puts-16                             33054550       36.26 ns/op       0 B/op       0 allocs/op
  Gets-16                            100000000       11.28 ns/op       0 B/op       0 allocs/op
  Mixed-16                            33174410       35.66 ns/op       0 B/op       0 allocs/op
    Hits : 10369406, Miss : 22805004, Ratio: 0.3126
  Mixed_Parallel-16                   58955106       19.96 ns/op       0 B/op       0 allocs/op

[ Random Workload ]
  Puts-16                             27217233       44.25 ns/op       0 B/op       0 allocs/op
  Gets-16                            100000000       11.55 ns/op       0 B/op       0 allocs/op
  Mixed-16                            32497040       37.01 ns/op       0 B/op       0 allocs/op
    Hits : 507588, Miss : 31989452, Ratio: 0.0156
  Mixed_Parallel-16                   77845632       15.53 ns/op       0 B/op       0 allocs/op

```

* `tlru.LRU` with `128` shards and `mux.NewX32` algorithm:

```text
[ Zipf Workload ]
  Puts-16                             33516466       35.88 ns/op       0 B/op       0 allocs/op
  Gets-16                            100000000       10.74 ns/op       0 B/op       0 allocs/op
  Mixed-16                            34404482       34.58 ns/op       0 B/op       0 allocs/op
    Hits : 10630761, Miss : 23773721, Ratio: 0.3090
  Mixed_Parallel-16                   60228430       19.66 ns/op       0 B/op       0 allocs/op

[ Random Workload ]
  Puts-16                             28159216       42.65 ns/op       0 B/op       0 allocs/op
  Gets-16                            100000000       10.83 ns/op       0 B/op       0 allocs/op
  Mixed-16                            34258018       33.99 ns/op       0 B/op       0 allocs/op
    Hits : 533000, Miss : 33725018, Ratio: 0.0156
  Mixed_Parallel-16                  100000000       11.66 ns/op       0 B/op       0 allocs/op

```

* `tlru.LRU` with `256` shards and `mux.NewX32` algorithm:

```text
[ Zipf Workload ]
  Puts-16                             34538334       34.77 ns/op       0 B/op       0 allocs/op
  Gets-16                            100000000       10.95 ns/op       0 B/op       0 allocs/op
  Mixed-16                            35645792       33.64 ns/op       0 B/op       0 allocs/op
    Hits : 10676624, Miss : 24969168, Ratio: 0.2995
  Mixed_Parallel-16                   75248469       16.22 ns/op       0 B/op       0 allocs/op

[ Random Workload ]
  Puts-16                             28202634       42.61 ns/op       0 B/op       0 allocs/op
  Gets-16                            100000000       11.24 ns/op       0 B/op       0 allocs/op
  Mixed-16                            35532368       33.46 ns/op       0 B/op       0 allocs/op
    Hits : 557243, Miss : 34975125, Ratio: 0.0157
  Mixed_Parallel-16                  121141429        9.90 ns/op       0 B/op       0 allocs/op

```

* `lrucore.Core` (Single-Threaded):

```text
[ Zipf Workload ]
  Puts-16                             28712246       41.79 ns/op       0 B/op       0 allocs/op
  Gets-16                            198613611        6.07 ns/op       0 B/op       0 allocs/op
  Mixed-16                            30072946       40.02 ns/op       0 B/op       0 allocs/op
    Hits : 9533449, Miss : 20539497, Ratio: 0.3170
  Mixed_Parallel-16                   11084142       91.50 ns/op       0 B/op       0 allocs/op

[ Random Workload ]
  Puts-16                             21257362       56.26 ns/op       0 B/op       0 allocs/op
  Gets-16                            193248979        6.14 ns/op       0 B/op       0 allocs/op
  Mixed-16                            25733350       45.89 ns/op       0 B/op       0 allocs/op
    Hits : 402293, Miss : 25331057, Ratio: 0.0156
  Mixed_Parallel-16                   13872712       85.06 ns/op       0 B/op       0 allocs/op
```

## License

Copyright(c) 2026 [Pranav R S](https://github.com/justpranavrs)

Licensed under [BSD-3-Clause](./LICENSE)

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
  - [How does lrucore.Core work?](#how-does-lrucorecore-work)
  - [What is tlru.LRU?](#what-is-tlrulru)
  - [What is a mux.Mux?](#what-is-a-muxmux)
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
- It has optimized batch operations like `GetMany` and `PutMany` which reduce the locking contention during high workloads. This is only limited to `lrucore.Core`.

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

### How does `TTL` work?
- Both `lrucore.Core` and `tlru.LRU` have a `WithTTL` option. Both of these instances use `Absolute TTL`. The timestamp of a `key` in the cache is updated only on `Put` operations and never on `Get` operations.

For a detailed walkthrough, refer [here](./LRU.md)

## Installation

```bash
go get -u github.com/justpranavrs/tlru@v0.5.0
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
cache, err := tlru.New[int, string](25600, tlru.WithTTL(5 * time.Hour))
```

#### **Note** : For more examples, refer [here](./lru_example_test.go)

## Benchmarks

### Environment

- os: archlinux/amd64
- cpu: AMD Ryzen 7 260 w/ Radeon 780M Graphics

### Performance

* `tlru.LRU` with `64` shards:
```text
[ Zipf Data ]
  Puts-16                             35025987       33.41 ns/op       0 B/op       0 allocs/op
  Gets-16                             50560573       23.90 ns/op       0 B/op       0 allocs/op
  Mixed-16                            40714761       29.79 ns/op       0 B/op       0 allocs/op
    Hits : 12709563, Miss : 28005198, Ratio: 0.3122
  Mixed_Parallel-16                   46865614       25.19 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             29462781       41.09 ns/op       0 B/op       0 allocs/op
  Gets-16                             64754400       18.42 ns/op       0 B/op       0 allocs/op
  Mixed-16                            32752305       34.82 ns/op       0 B/op       0 allocs/op
    Hits : 512526, Miss : 32239779, Ratio: 0.0156
  Mixed_Parallel-16                   76516498       15.67 ns/op       0 B/op       0 allocs/op
```

* `tlru.LRU` with `128` shards (Default):
```text
[ Zipf Data ]
  Puts-16                             33495073       33.87 ns/op       0 B/op       0 allocs/op
  Gets-16                             48900223       23.23 ns/op       0 B/op       0 allocs/op
  Mixed-16                            38118705       30.13 ns/op       0 B/op       0 allocs/op
    Hits : 11726697, Miss : 26392008, Ratio: 0.3076
  Mixed_Parallel-16                   59304232       19.36 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             27993507       40.39 ns/op       0 B/op       0 allocs/op
  Gets-16                             72939033       16.45 ns/op       0 B/op       0 allocs/op
  Mixed-16                            35675577       33.37 ns/op       0 B/op       0 allocs/op
    Hits : 555180, Miss : 35120397, Ratio: 0.0156
  Mixed_Parallel-16                   93977918       12.18 ns/op       0 B/op       0 allocs/op
```

* `tlru.LRU` with `256` shards:
```text
[ Zipf Data ]
  Puts-16                             34796096       34.50 ns/op       0 B/op       0 allocs/op
  Gets-16                             50799690       23.18 ns/op       0 B/op       0 allocs/op
  Mixed-16                            39515062       30.38 ns/op       0 B/op       0 allocs/op
    Hits : 11794110, Miss : 27720952, Ratio: 0.2985
  Mixed_Parallel-16                   72371682       16.57 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             29310763       40.49 ns/op       0 B/op       0 allocs/op
  Gets-16                             71429926       16.24 ns/op       0 B/op       0 allocs/op
  Mixed-16                            36201361       33.51 ns/op       0 B/op       0 allocs/op
    Hits : 562800, Miss : 35638561, Ratio: 0.0155
  Mixed_Parallel-16                  100000000       10.39 ns/op       0 B/op       0 allocs/op
```

* `lrucore.Core` (Single-Threaded):
```text
[ Zipf Data ]
  Puts-16                             31053111       38.15 ns/op       0 B/op       0 allocs/op
  Gets-16                             48959450       23.88 ns/op       0 B/op       0 allocs/op
  Mixed-16                            33959961       32.46 ns/op       0 B/op       0 allocs/op
    Hits : 10767981, Miss : 23191980, Ratio: 0.3171
  Mixed_Parallel-16                   14762529       79.64 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             21556010       56.26 ns/op       0 B/op       0 allocs/op
  Gets-16                             47074131       21.35 ns/op       0 B/op       0 allocs/op
  Mixed-16                            27001513       45.61 ns/op       0 B/op       0 allocs/op
    Hits : 421099, Miss : 26580414, Ratio: 0.0156
  Mixed_Parallel-16                   12744640       94.30 ns/op       0 B/op       0 allocs/op
```

## License

Copyright(c) 2026 [Pranav R S](https://github.com/justpranavrs)

Licensed under [BSD-3-Clause](./LICENSE)

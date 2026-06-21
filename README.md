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
- cpu: AMD Ryzen 7 260 w/ Radeon 780M Graphics

### Performance

* `tlru.LRU` with `64` shards:
```text
[ Zipf Data ]
  Puts-16                             36164115       34.87 ns/op       0 B/op       0 allocs/op
  Gets-16                             97302411       13.20 ns/op       0 B/op       0 allocs/op
  Mixed-16                            34324671       35.01 ns/op       0 B/op       0 allocs/op
    Hits : 10739429, Miss : 23585242, Ratio: 0.3129
  Mixed_Parallel-16                   53400261       22.07 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             29226392       41.45 ns/op       0 B/op       0 allocs/op
  Gets-16                             81335832       13.05 ns/op       0 B/op       0 allocs/op
  Mixed-16                            34635925       34.57 ns/op       0 B/op       0 allocs/op
    Hits : 539935, Miss : 34095990, Ratio: 0.0156
  Mixed_Parallel-16                   76091194       16.21 ns/op       0 B/op       0 allocs/op
```

* `tlru.LRU` with `128` shards (Default):
```text
[ Zipf Data ]
  Puts-16                             35838956       35.16 ns/op       0 B/op       0 allocs/op
  Gets-16                             91117611       13.18 ns/op       0 B/op       0 allocs/op
  Mixed-16                            32812344       35.68 ns/op       0 B/op       0 allocs/op
    Hits : 10086308, Miss : 22726036, Ratio: 0.3074
  Mixed_Parallel-16                   56536936       19.55 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             29004740       42.09 ns/op       0 B/op       0 allocs/op
  Gets-16                             73972798       13.77 ns/op       0 B/op       0 allocs/op
  Mixed-16                            32143723       35.59 ns/op       0 B/op       0 allocs/op
    Hits : 503441, Miss : 31640282, Ratio: 0.0157
  Mixed_Parallel-16                   97455319       11.89 ns/op       0 B/op       0 allocs/op
```

* `tlru.LRU` with `256` shards:
```text
[ Zipf Data ]
  Puts-16                             33154726       35.86 ns/op       0 B/op       0 allocs/op
  Gets-16                             82692465       14.05 ns/op       0 B/op       0 allocs/op
  Mixed-16                            34448319       34.63 ns/op       0 B/op       0 allocs/op
    Hits : 10296718, Miss : 24151601, Ratio: 0.2989
  Mixed_Parallel-16                   65244442       17.40 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             27685598       44.87 ns/op       0 B/op       0 allocs/op
  Gets-16                             82208396       14.34 ns/op       0 B/op       0 allocs/op
  Mixed-16                            33635023       37.73 ns/op       0 B/op       0 allocs/op
    Hits : 525149, Miss : 33109874, Ratio: 0.0156
  Mixed_Parallel-16                  100000000       10.10 ns/op       0 B/op       0 allocs/op
```

* `lrucore.Core`:
```text
[ Zipf Data ]
  Puts-16                             30797480       37.65 ns/op       0 B/op       0 allocs/op
  Gets-16                            169850612        7.11 ns/op       0 B/op       0 allocs/op
  Mixed-16                            33556694       36.07 ns/op       0 B/op       0 allocs/op
    Hits : 10637908, Miss : 22918786, Ratio: 0.3170
  Mixed_Parallel-16                   15446341       85.77 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             21273944       55.90 ns/op       0 B/op       0 allocs/op
  Gets-16                            173406368        6.86 ns/op       0 B/op       0 allocs/op
  Mixed-16                            25647774       45.42 ns/op       0 B/op       0 allocs/op
    Hits : 400380, Miss : 25247394, Ratio: 0.0156
  Mixed_Parallel-16                   14873089       93.58 ns/op       0 B/op       0 allocs/op
```

## License

Copyright(c) 2026 [Pranav R S](https://github.com/justpranavrs)

Licensed under [BSD-3-Clause](./LICENSE)

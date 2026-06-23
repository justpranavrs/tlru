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
  Puts-16                             36523065       33.26 ns/op       0 B/op       0 allocs/op
  Gets-16                             49998720       24.27 ns/op       0 B/op       0 allocs/op
  Mixed-16                            38868608       30.42 ns/op       0 B/op       0 allocs/op
    Hits : 12198270, Miss : 26670338, Ratio: 0.3138
  Mixed_Parallel-16                   58957501       20.42 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             27762294       42.16 ns/op       0 B/op       0 allocs/op
  Gets-16                             60976142       18.82 ns/op       0 B/op       0 allocs/op
  Mixed-16                            32838690       36.67 ns/op       0 B/op       0 allocs/op
    Hits : 512798, Miss : 32325892, Ratio: 0.0156
  Mixed_Parallel-16                   72249718       16.89 ns/op       0 B/op       0 allocs/op
```

* `tlru.LRU` with `128` shards (Default):
```text
[ Zipf Data ]
  Puts-16                             32705001       34.91 ns/op       0 B/op       0 allocs/op
  Gets-16                             44981125       25.64 ns/op       0 B/op       0 allocs/op
  Mixed-16                            36485504       33.10 ns/op       0 B/op       0 allocs/op
    Hits : 11273690, Miss : 25211814, Ratio: 0.3090
  Mixed_Parallel-16                   59793296       20.23 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             27287380       43.57 ns/op       0 B/op       0 allocs/op
  Gets-16                             62473988       17.91 ns/op       0 B/op       0 allocs/op
  Mixed-16                            29104575       36.34 ns/op       0 B/op       0 allocs/op
    Hits : 454279, Miss : 28650296, Ratio: 0.0156
  Mixed_Parallel-16                   88131531       14.40 ns/op       0 B/op       0 allocs/op
```

* `tlru.LRU` with `256` shards:
```text
[ Zipf Data ]
  Puts-16                             31912034       36.32 ns/op       0 B/op       0 allocs/op
  Gets-16                             48343602       25.05 ns/op       0 B/op       0 allocs/op
  Mixed-16                            37630839       31.63 ns/op       0 B/op       0 allocs/op
    Hits : 11223834, Miss : 26407005, Ratio: 0.2983
  Mixed_Parallel-16                   68040799       17.54 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             26496339       43.42 ns/op       0 B/op       0 allocs/op
  Gets-16                             71081203       16.68 ns/op       0 B/op       0 allocs/op
  Mixed-16                            33968317       33.77 ns/op       0 B/op       0 allocs/op
    Hits : 530169, Miss : 33438148, Ratio: 0.0156
  Mixed_Parallel-16                  100000000       10.61 ns/op       0 B/op       0 allocs/op
```

* `lrucore.Core`:
```text
[ Zipf Data ]
  Puts-16                             31436680       37.80 ns/op       0 B/op       0 allocs/op
  Gets-16                             49003570       24.76 ns/op       0 B/op       0 allocs/op
  Mixed-16                            34149271       33.48 ns/op       0 B/op       0 allocs/op
    Hits : 10827918, Miss : 23321353, Ratio: 0.3171
  Mixed_Parallel-16                   16186285       76.69 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             20627521       55.77 ns/op       0 B/op       0 allocs/op
  Gets-16                             47857557       21.62 ns/op       0 B/op       0 allocs/op
  Mixed-16                            26897581       45.96 ns/op       0 B/op       0 allocs/op
    Hits : 418846, Miss : 26478735, Ratio: 0.0156
  Mixed_Parallel-16                   14360539       83.78 ns/op       0 B/op       0 allocs/op
```

## License

Copyright(c) 2026 [Pranav R S](https://github.com/justpranavrs)

Licensed under [BSD-3-Clause](./LICENSE)

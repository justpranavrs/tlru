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
- `tlru.TLRU` and `lrucore.TTLCore` are TTL implementations of `tlru.LRU` and `lrucore.Core` respectively. They use `Absolute TTL`. The timestamp of a `key` in the cache is updated only on `Put` operations and never on `Get` operations.
- `WithSliding` on these instances enable `Sliding TTL` instead of the default `Absolute TTL`. `Sliding TTL` ensures timestamp updates on `Get` and `Peek` operations too.

For a detailed walkthrough, refer [here](./LRU.md)

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

	"github.com/justpranavrs/tlru"
)

func main() {
	// create a new TLRU cache instance.
  // default number of containers is 128.
	cache, err := tlru.NewTTL[int, int](51200, 24 * time.Hour)
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

* `tlru.LRU` with `64` shards:
```text
[ Zipf Data ]
  Puts-16                             34108743       35.38 ns/op       0 B/op       0 allocs/op
  Gets-16                             48408013       24.90 ns/op       0 B/op       0 allocs/op
  Mixed-16                            38746975       31.22 ns/op       0 B/op       0 allocs/op
    Hits : 12153545, Miss : 26593430, Ratio: 0.3137
  Mixed_Parallel-16                   55047906       21.60 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             27639721       43.27 ns/op       0 B/op       0 allocs/op
  Gets-16                             62132617       19.13 ns/op       0 B/op       0 allocs/op
  Mixed-16                            31619710       36.55 ns/op       0 B/op       0 allocs/op
    Hits : 494969, Miss : 31124741, Ratio: 0.0157
  Mixed_Parallel-16                   72448735       16.64 ns/op       0 B/op       0 allocs/op
```

* `tlru.LRU` with `128` shards (Default):
```text
[ Zipf Data ]
  Puts-16                             33837464       36.00 ns/op       0 B/op       0 allocs/op
  Gets-16                             49723561       24.15 ns/op       0 B/op       0 allocs/op
  Mixed-16                            37874719       31.43 ns/op       0 B/op       0 allocs/op
    Hits : 11666386, Miss : 26208333, Ratio: 0.3080
  Mixed_Parallel-16                   54515466       21.72 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             28234916       42.82 ns/op       0 B/op       0 allocs/op
  Gets-16                             66785313       17.37 ns/op       0 B/op       0 allocs/op
  Mixed-16                            33921026       35.20 ns/op       0 B/op       0 allocs/op
    Hits : 530527, Miss : 33390499, Ratio: 0.0156
  Mixed_Parallel-16                   88540772       13.73 ns/op       0 B/op       0 allocs/op
```

* `tlru.LRU` with `256` shards:
```text
[ Zipf Data ]
  Puts-16                             32998448       35.51 ns/op       0 B/op       0 allocs/op
  Gets-16                             49472197       23.82 ns/op       0 B/op       0 allocs/op
  Mixed-16                            36803145       31.52 ns/op       0 B/op       0 allocs/op
    Hits : 10993870, Miss : 25809275, Ratio: 0.2987
  Mixed_Parallel-16                   72907016       16.58 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             26257032       43.70 ns/op       0 B/op       0 allocs/op
  Gets-16                             72060157       16.97 ns/op       0 B/op       0 allocs/op
  Mixed-16                            34110064       35.29 ns/op       0 B/op       0 allocs/op
    Hits : 533155, Miss : 33576909, Ratio: 0.0156
  Mixed_Parallel-16                  100000000       11.98 ns/op       0 B/op       0 allocs/op
```

* `lrucore.Core` (Single-Threaded):
```text
[ Zipf Data ]
  Puts-16                             28654330       39.76 ns/op       0 B/op       0 allocs/op
  Gets-16                             53279278       22.59 ns/op       0 B/op       0 allocs/op
  Mixed-16                            36166560       33.42 ns/op       0 B/op       0 allocs/op
    Hits : 11467612, Miss : 24698948, Ratio: 0.3171
  Mixed_Parallel-16                   15649838       76.10 ns/op       0 B/op       0 allocs/op

[ Random Data ]
  Puts-16                             20790658       57.83 ns/op       0 B/op       0 allocs/op
  Gets-16                             58139402       20.40 ns/op       0 B/op       0 allocs/op
  Mixed-16                            25419280       46.27 ns/op       0 B/op       0 allocs/op
    Hits : 396322, Miss : 25022958, Ratio: 0.0156
  Mixed_Parallel-16                   15199792       80.98 ns/op       0 B/op       0 allocs/op
```

## License

Copyright(c) 2026 [Pranav R S](https://github.com/justpranavrs)

Licensed under [BSD-3-Clause](./LICENSE)

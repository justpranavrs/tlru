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
- It uses a `mux.MuxHash` to route the key to one of its shards.
- It has two options: 
  - `WithShards`: It allows the user to customize the number of shards `tlru.LRU` creates.
  - `WithMux`: It allows the configuration of `mux.MuxHash`.

For a detailed walkthrough, refer [here](./LRU.md)

## Installation

```bash
go get -u github.com/justpranavrs/tlru@latest
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

```
goos: linux
goarch: amd64
pkg: github.com/justpranavrs/tlru
cpu: AMD Ryzen 7 260 w/ Radeon 780M Graphics

BenchmarkLRUWith64/Zipf_Puts-16                     33895142      35.50 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Zipf_Gets-16                    100000000      10.74 ns/op       0 B/op       0 allocs/op
Hits : 10346504, Miss : 22689820, Ratio: 0.3132

BenchmarkLRUWith64/Zipf_Mixed-16                    33036324      34.23 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Zipf_Mixed_Parallel-16           59677258      19.84 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Random_Puts-16                   28279300      42.84 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Random_Gets-16                  100000000      11.16 ns/op       0 B/op       0 allocs/op
Hits : 530268, Miss : 33483160, Ratio: 0.0156

BenchmarkLRUWith64/Random_Mixed-16                  34013428      34.95 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith64/Random_Mixed_Parallel-16         76379761      15.54 ns/op       0 B/op       0 allocs/op

BenchmarkLRU/Zipf_Puts-16                           32656112      36.36 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Zipf_Gets-16                          100000000      11.24 ns/op       0 B/op       0 allocs/op
Hits : 10665195, Miss : 23736553, Ratio: 0.3100

BenchmarkLRU/Zipf_Mixed-16                          34401748      34.94 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Zipf_Mixed_Parallel-16                 60444584      19.27 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Random_Puts-16                         28685742      41.55 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Random_Gets-16                        100000000      11.01 ns/op       0 B/op       0 allocs/op
Hits : 565990, Miss : 35638611, Ratio: 0.0156

BenchmarkLRU/Random_Mixed-16                        36204601      33.29 ns/op       0 B/op       0 allocs/op
BenchmarkLRU/Random_Mixed_Parallel-16              100000000      11.78 ns/op       0 B/op       0 allocs/op

BenchmarkLRUWith256/Zipf_Puts-16                    34739044      37.05 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Zipf_Gets-16                    90376778      11.86 ns/op       0 B/op       0 allocs/op
Hits : 8720847, Miss : 20926643, Ratio: 0.2942

BenchmarkLRUWith256/Zipf_Mixed-16                   29647490      37.26 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Zipf_Mixed_Parallel-16          73194765      16.23 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Random_Puts-16                  28679670      41.81 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Random_Gets-16                 100000000      11.42 ns/op       0 B/op       0 allocs/op
Hits : 556574, Miss : 35059573, Ratio: 0.0156

BenchmarkLRUWith256/Random_Mixed-16                 35616147      32.98 ns/op       0 B/op       0 allocs/op
BenchmarkLRUWith256/Random_Mixed_Parallel-16       100000000      10.06 ns/op       0 B/op       0 allocs/op
PASS
ok      github.com/justpranavrs/tlru    117.784s

---

goos: linux
goarch: amd64
pkg: github.com/justpranavrs/tlru/lrucore
cpu: AMD Ryzen 7 260 w/ Radeon 780M Graphics

BenchmarkLRUCore/Zipf_Puts-16                       28624830      42.47 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Zipf_Gets-16                      176645416       6.66 ns/op       0 B/op       0 allocs/op
Hits : 9701277, Miss : 20902822, Ratio: 0.3170

BenchmarkLRUCore/Zipf_Mixed-16                      30604099      39.83 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Zipf_Mixed_Parallel-16             12427209     100.40 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Random_Puts-16                     18420439      59.85 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Random_Gets-16                    173226816       6.78 ns/op       0 B/op       0 allocs/op
Hits : 381034, Miss : 24014774, Ratio: 0.0156

BenchmarkLRUCore/Random_Mixed-16                    24395808      48.63 ns/op       0 B/op       0 allocs/op
BenchmarkLRUCore/Random_Mixed_Parallel-16           13277698     101.00 ns/op       0 B/op       0 allocs/op
PASS
ok      github.com/justpranavrs/tlru/lrucore    40.171s
```

## License
Copyright(c) 2026 [Pranav R S](https://github.com/justpranavrs)

Licensed under [BSD-3-Clause](./LICENSE)
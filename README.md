# tlru

**tlru** is a high-performance, array based, time-aware **LRU** cache for Go with **zero runtime allocations** and **zero dependencies**. It also supports utilizing multiple independent containers to eliminate lock contention and allow high-concurrency operations without bottlenecks.

#### **NOTE**: The current version has no support for TTL. It will be added in the future versions.

## Table of Contents
- [Installation](#installation)
- [Benchmarks](#benchmarks)
- [Examples](#examples)
    - [Basic LRU Cache](#basic-lru-cache)
    - [Customization](#customization)
- [License](#license)

## Installation

```bash
go get -u github.com/justpranavrs/tlru
```

## Benchmarks
| Component & Workload | Iterations | Latency | Memory | Allocations |
| --- | --- | --- | --- | --- |
| **`tlru` (Basic LRU Cache)** |  |  |  |  |
| Zipf Puts | 23,886,063 | 55.05 ns/op | 10 B/op | 1 allocs/op |
| Zipf Gets | 42,368,697 | 30.47 ns/op | 10 B/op | 1 allocs/op |
| Zipf Mixed | 23,678,805 | 52.33 ns/op | 10 B/op | 1 allocs/op |
| **Zipf Mixed Parallel** | **61,236,249** | **19.19 ns/op** | **10 B/op** | **1 allocs/op** |
| Random Puts | 18,772,122 | 62.63 ns/op | 15 B/op | 1 allocs/op |
| Random Gets | 40,221,781 | 29.62 ns/op | 15 B/op | 1 allocs/op |
| Random Mixed | 20,971,242 | 57.95 ns/op | 15 B/op | 1 allocs/op |
| **Random Mixed Parallel** | **72,212,966** | **15.43 ns/op** | **15 B/op** | **1 allocs/op** |
|  |  |  |  |  |
| **`tlru` (Zero Allocation, Sharded LRU Cache)** |  |  |  |  |
| Zipf Puts | 27,370,032 | 38.32 ns/op | 0 B/op | 0 allocs/op |
| Zipf Gets | 82,687,224 | 14.10 ns/op | 0 B/op | 0 allocs/op |
| Zipf Mixed | 31,359,452 | 38.13 ns/op | 0 B/op | 0 allocs/op |
| **Zipf Mixed Parallel** | **63,528,969** | **17.29 ns/op** | **0 B/op** | **0 allocs/op** |
| Random Puts | 24,723,649 | 45.68 ns/op | 0 B/op | 0 allocs/op |
| Random Gets | 80,985,068 | 14.77 ns/op | 0 B/op | 0 allocs/op |
| Random Mixed | 32,517,832 | 38.02 ns/op | 0 B/op | 0 allocs/op |
| **Random Mixed Parallel** | **87,168,492** | **14.36 ns/op** | **0 B/op** | **0 allocs/op** |
|  |  |  |  |  |
| **`lrucore` (Single Threaded LRU Cache)** |  |  |  |  |
| Zipf Puts | 27,305,832 | 43.23 ns/op | 0 B/op | 0 allocs/op |
| Zipf Gets | 172,443,609 | 6.926 ns/op | 0 B/op | 0 allocs/op |
| Zipf Mixed | 28,464,217 | 41.26 ns/op | 0 B/op | 0 allocs/op |
| Zipf Mixed Parallel | 11,175,440 | 95.03 ns/op | 0 B/op | 0 allocs/op |
| Random Puts | 20,641,852 | 58.40 ns/op | 0 B/op | 0 allocs/op |
| Random Gets | 182,889,151 | 6.570 ns/op | 0 B/op | 0 allocs/op |
| Random Mixed | 25,888,710 | 46.57 ns/op | 0 B/op | 0 allocs/op |
| Random Mixed Parallel | 13,382,323 | 96.44 ns/op | 0 B/op | 0 allocs/op |

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
cache, err := tlru.New(cacheCapacity, tlru.WithUnsafe[int, User]())
```

#### **Note** : For more examples, refer [here](./lru_example_test.go)
## License
Copyright(c) 2026 Pranav R S

Licensed under [BSD-3-Clause](./LICENSE)
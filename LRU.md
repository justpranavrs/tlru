## Introduction

Hello, This document is designed to help you get started with `tlru` and how to use `tlru.PoolLRU` to its full capacity.

## Table of Contents

- [Getting Started](#getting-started)
- [Customizing the Number of Shards](#customizing-the-number-of-shards)
- [Customizing the Mux Algorithm](#customizing-the-mux-algorithm)
  - [Using the `tlru/mux` package](#using-the-tlrumux-package)
  - [Using a customized mux algorithm](#using-a-customized-mux-algorithm)
- [Enabling TTL(Time-To-Live)](#enabling-ttltime-to-live)
  - [Using Sliding TTL](#using-sliding-ttl)
  - [Using a Custom Clock](#using-a-custom-clock)

### Getting Started

This is a detailed walkthrough on how to get started with `tlru`.

The choice of using either `tlru.PoolLRU` or `core.LRU` depends on

- `tlru.PoolLRU` works on `shard-based local eviction`. It consists of multiple `core.LRU` instances known as `shards`. It does not care about the globally oldest key. While this does go around the textbook definition of LRU Cache, in practical cases, it gives `higher performance on high concurrency workloads` compared to its parent. It is less limited to mutual extension locks, because of its `sharded architecture`. For more details on how sharded architectures work, refer `database sharding` [here](https://www.geeksforgeeks.org/system-design/database-sharding-a-system-design-concept/).
- Its parent, `core.LRU` works on the pure LRU Cache definition, it evicts the `globally oldest key`. It is only useful in scenarios where this matters. It performs a bit slower because of mutual extension locks, `sync.Mutex` for majority of its operations.

A simple `tlru.PoolLRU` instance can be created using the `tlru.New` constructor. It takes in the cache capacity as its argument.

```go
cache, err := tlru.New[int, int](25600)
```

The `[int, int]` is use of Go Generics, introduced in `Go 1.18`. Refer [here](https://go.dev/doc/tutorial/generics).

The above instance has a default of `128` shards, distributed with a capacity of 200 each, `25600 / 128 = 200`.

If the capacity is not perfectly divisible by the number of shards, the remainder will go to some of the shards, leaving an uneven distribution. It is recommended to provide the number of shards as a factor of the capacity for `even distributions`.

### Customizing the Number of Shards

To customize the number of shards, the `WithShards` method has to be used as shown below.

```go
cache, err := tlru.New[int, int](25600, tlru.WithShards(64))
```

The above snippet creates `64` instances instead of 128, by distributing 400 capacity to each instance.

**NOTE**: Increasing the number of shards will result in better speed but at the cost of losing the core feature of LRU. It will lead to immature evictions.

### Customizing the Mux Algorithm

#### Using the `tlru/mux` package

To customize the mux hashing algorithm, the `WithMux` method has to be used as shown below.

```go
capacity := 25600
shards := 128
cache, err := tlru.New[int, string](capacity, tlru.WithMux(mux.NewF32[int](shards)))
```

The above snippet uses the FNV-1a algorithm, than the default `hash/maphash` algorithm.
Below are the given algorithms currently in the `tlru/mux` package: 
  - `FNV-1a` 
  - `xxHash32` 
  - `hash/maphash`

**NOTE**: The `hash/maphash` implementation, which is `mux.NewMH32` is compatible with all key types of type comparable, while the other two implementations lack support for floats and custom structs.

#### Using a customized mux algorithm

To use a custom mux hash algorithm, it has to be of type `mux.Mux` which is,

```go
type Mux[K comparable] func(key K) uint32
```

Here is a basic implementation of a Mux

```go
nShards := 128

// The hashing algorithm must return a number of uint32 type between 0 and nShards-1.
func CustomMux[K comparable](key int) uint32 {
	return bits.RotateLeft32(uint32(key) & (nShards-1), 16)
}
```

As you can clearly see, the above snippet is a terrible example for a custom hash algorithm, it works but at the same time, it is vulnerable to Hash-DOS attacks. But it demonstrates the example of how to create a `mux.Mux` to be able to use it in `tlru.PoolLRU`

```go
cache, err := tlru.New[int, string](25600, tlru.WithMux(CustomMux[int]))
```

### Enabling TTL(Time-To-Live)
The `tlru.PoolTLRU` and `core.TLRU` instances are the TTL implementations of `tlru.PoolLRU` and `core.TLRU` respectively. It uses `Absolute TTL`, which does not update the timestamp for the `key` during a `Get` operation.

The below examples demonstrates how to create a cache with a `TTL` of `5 hours`.
```go
cache, err := tlru.NewWithTTL[int, string](25600, 5 * time.Hour)
```

When a cache is created, a background clock, which is a goroutine is spawned. To safely close the goroutine, calling `cache.Close()` is the best and recommended practice.

For a single instance `core.LRU` with TTL, `core.TLRU` is available, and it can be created using `core.NewWithTTL`.
```go
cache, err := core.NewWithTTL[int, string](25600, 5 * time.Hour)
```

#### Using Sliding TTL
The `WithSliding` option enables `Sliding TTL` which updates the timestamps of the keys during `Get` and `Peek` operations too.

The below examples demonstrates how to create a cache with `Sliding TTL`.
```go
cache, err := tlru.NewWithTTL[int, string](25600, 5 * time.Hour, tlru.WithSliding())
```

#### Using a Custom Clock
LRU Cache with TTL uses a background clock instead of the CPU's clock to reduce the lock contention due to `sync/Mutex` by using heavy operations inside a lock.

The default clock duration is 100ms. To customize it, the `tlru/clock` package is used.
```go
clock := clock.New(200 * time.Millisecond)
cache, err := tlru.NewWithTTL[int, int](25600, 5 * time.Hour, tlru.WithClock(clock))
```
Above example uses a clock with 200ms.

You can look at more examples [here](./lru_example_test.go)

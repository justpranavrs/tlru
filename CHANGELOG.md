# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/2.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2026-06-24
Link : [Github Release](https://github.com/justpranavrs/tlru/releases/tag/v0.5.0)

### Added
- `TTL` can be enabled using the `WithTTL` option for both `lrucore.Core` and `tlru.LRU` and it uses `Absolute TTL`.
- `Delete` is now available to both `tlru.LRU` and `lrucore.Core`. It returns `false` if they key was not present in the cache, else it returns true and also the evicted value. Also an Example for `Delete` was added.
- `tlru/lruclock` which allows creating background clocks for TTL with a custom timer.
- `Close` to `tlru.Cache` and both `tlru.LRU` and `lrucore.Core` to safely close the background clock.

### Changed
- `lrucore.Core` internal architecture has been changed. The entire cache is always linked, embedding the `free` pointer doubly-linked list just after the `mru` of the cache. This approach was taken to allow `Delete` and `TTL`.

## [0.4.1] - 2026-06-23
Link : [Github Release](https://github.com/justpranavrs/tlru/releases/tag/v0.4.1)

### Changed
- `WithMux` now raises compile-time error, rather than runtime error for `NewF32` and `NewX32`.
- `NewF32`, `NewX32` now only return `mux.Mux` instead of `(mux.Mux, error)`.
- `FuzzCache` in `internal/testutil` now directly takes in `mux.Mux` instead of its standalone `TestMux`.

### Fixed
- `Benchmark_Gets` now retains the cache from `Benchmark_Puts`.
- `TestRaceCache` for `string` were using `uint`.
- `WithShards` boundary conditions more strict to ensure that the error is caught by `ErrInvalidShards` instead of `ErrInvalidCapacity`.

## [0.4.0] - 2026-06-21
Link : [Github Release](https://github.com/justpranavrs/tlru/releases/tag/v0.4.0)

### Added
- Replace `PutGrew` on `tlru.Cache`, `tlru.LRU` and `lrucore.Core` with `Upsert` which returns one of three states of `lrucore.UpsertState`:
    - `AddNoEvict` if a new key was added without eviction.
    - `AddOnEvict` if a older key was evicted and a new key was added.
    - `Replace` if an older key's value was replaced in the cache.
- `GetMany` and `PutMany` to `lrucore.Core` which enables batch operations under a single internal `mutex` lock.
- `Stats` to `tlru.Cache`, `tlru.LRU` and `lrucore.Core` to accurately measure the metrics of the instance. It can be reset using `ResetStats`.
    - `Hits`
    - `Misses`
    - `Evictions`
- `lrucore.Core.Shards` which always returns 1. 
- `tlru.Cache` also has the `Shards` method.
- `CHANGELOG.md`

### Changed
- Rename `lrucore.LRUCore` to `lrucore.Core`.
- `lrucore.Core` has been refactored internally to make all of the operations faster by separating `nodes` into two separate arrays.
- Default `mux.Mux` for `tlru.LRU` has been changed to `mux.NewMH32`
- The errors from `internal/errs` have moved to their respective packages for public access to allow for the `errors.Is` check.
- Optimized `Flush` to use Go's native `clear` for `lrucore.Core` and `tlru.LRU`.
- Make `internal/testutil` more organized.

### Removed
- `Compaction` and `Contains` from `tlru.Cache`, `tlru.LRU` and `lrucore.Core`. `Contains` can comfortably be replaced by `Peek`.

### Fixed
- `tlru.New` incorrectly allocating capacity to each of the shards when `capacity` was not divisible by `shards` after `v0.3.0`.
- `WithShards` to raise `tlru.ErrInvalidShards` when `num` exceeds `int32` range after `v0.3.0`.
- `lrucore.Core.Size` by adding `sync.Mutex` locks to prevent data race conditions.
- `mux.ErrInvalidMuxF32` and `mux.ErrInvalidMuxX32` from giving an incorrect not supported `int64` and `uint64` messages.

## [0.3.2] - 2026-06-19
Link : [Github Release](https://github.com/justpranavrs/tlru/releases/tag/v0.3.2)

### Added
- `README.md` badges.
- `Race` tests for `int32`, `uint`, `string`.

### Fixed
- `mux.NewX32` data race condition for `strings` with more than `16` characters.

## [0.3.1] - 2026-06-19
Link : [Github Release](https://github.com/justpranavrs/tlru/releases/tag/v0.3.1)

### Changed
- `tlru` support is updated from Go 1.22+ to 1.24+ due to the introduction of `hash/maphash`.
- `mux.MuxNumber` is now private and cannot be accessed by external packages.

## [0.3.0] - 2026-06-19
Link : [Github Release](https://github.com/justpranavrs/tlru/releases/tag/v0.3.0)

### Added
- `mux.NewMH32` uses `hash/maphash` to quickly hash non-primitives, by replacing the slow `encoding/json`.
- `tlru.LRU` has a new `Shards` method which reports the number of shards in the `LRU` instance. This method is not available in the `tlru.Cache` interface.
- If a key of type `float`, `complex` or `struct` is passed to the default `tlru.New` constructor without custom mux implementations, it will default from `mux.NewX32` to `mux.NewMH32` automatically.
- `Race` tests with goroutines.

### Changed
- `mux.MuxF32`, `mux.MuxX32` and `mux.MuxHash` have been removed and instead replaced with a singular `mux.Mux`.
- `mux.Mux` takes significantly faster due to the initial allocation of the hash function based on the type of key on `tlru.LRU`, eliminating the need of `switch-case` during runtime.
- `tlru.WithShards` no longer rounds its argument up to the next power of 2. This allows `tlru.LRU` to configure a non-power of 2 number of shards.
- `mux.NewF32` and `mux.NewX32` have no longer support for `float`, `complex` and `structs`. They can only be used with `string`, `bool`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr`.
- `tlru.Cache.GetQuiet` has been renamed to `tlru.Cache.Peek`.
- `tlru.Cache.PutGrows` has been renamed to `tlru.Cache.PutGrew`.
- Advanced CI workflow in Github Actions.

### Fixed
- `Fuzz` tests not using a source of truth for `PutGrew`.

### Security
- `mux.NewX32` which was prone to length-extension attacks for key with any numbers.

## [0.2.0] - 2026-06-18
Link : [Github Release](https://github.com/justpranavrs/tlru/releases/tag/v0.2.0)

### Added
- `tlru/mux` package, supports custom hash algorithms.
- `xxHash32` to `tlru/mux`.
- `tlru.LRU.PutGrows` and `lrucore.LRUCore.PutGrows` which returns true if size of the cache has been increased.
- `LRU.md` walkthrough for users.
- `Zipf` data tests for Benchmarks.
- Detailed `Benchmarks` in `README.md`.

### Changed
- `unsafe` package has been eliminated achieving zero-allocations without it.
- `tlru.LRU.Size` now takes O(1) instead of O(shards).
- `mux.Mux32` has been renamed to `mux.MuxF32` to prevent naming conflicts.

### Removed
- `LRUOption` has been removed and replaced with `Option` for cleaner API calls.
- `mux.MuxF32.fnvString`, string passed as `UTF-8`.
- `mux.MuxF32.Get` with bool when number of shards is 1.
- `internal/conv` package removed.

### Fixed
- `lrucore.LRUCore.Size` data race condition due to mixed usage of `sync/atomic` and `int32`.

## [0.1.1] - 2026-06-18
Link : [Github Release](https://github.com/justpranavrs/tlru/releases/tag/v0.1.1)

### Added
- `Introduction` and `Benchmarks` in `README.md`.
- More `Examples` for GoDoc.

### Fixed
- `tlru.New` now returns `*tlru.LRU[K, V]` instead of `tlru.Cache[K, V]`.

## [0.1.0] - 2026-06-18
Link : [Github Release](https://github.com/justpranavrs/tlru/releases/tag/v0.1.0)

### Added
- `lruCore.LRUCore`, an array-based doubly-linked list implementation of the `Least Recently Used` cache. It guarantees zero runtime allocations.
- Add `sync.Mutex` to `lrucore.LRUCore` to avoid concurrency issues.
- `tlru.LRU`, a sharded instance of `lrucore.LRUCore` with a default of 128 instances. It uses `mux.Mux32` with custom-offset `FNV-1a` to prevent Hash-DOS attacks. It uses shard-based local eviction rather than global eviction.
- `WithShards` and `WithUnsafe`, two options to configure `tlru.LRU`. 
- Go 1.18+ `Generics` support.
- `tlru.Cache`, an interface defining both `tlru.LRU` and `lrucore.LRUCore`:
    - `Capacity`
    - `Compaction`
    - `Contains`
    - `Flush`
    - `Get`
    - `GetQuiet`
    - `Put`
    - `Size`
- CI workflow for Github Actions.
- `Table-Driven` unit tests, `Fuzz` tests and `Benchmark` tests.
- `Examples` for GoDoc.

## 
- [0.5.0] : [View changes from 0.4.1 to 0.5.0](https://github.com/justpranavrs/tlru/compare/v0.4.1...v0.5.0)
- [0.4.1] : [View changes from 0.4.0 to 0.4.1](https://github.com/justpranavrs/tlru/compare/v0.4.0...v0.4.1)
- [0.4.0] : [View changes from 0.3.2 to 0.4.0](https://github.com/justpranavrs/tlru/compare/v0.3.2...v0.4.0)
- [0.3.2] : [View changes from 0.3.1 to 0.3.2](https://github.com/justpranavrs/tlru/compare/v0.3.1...v0.3.2)
- [0.3.1] : [View changes from 0.3.0 to 0.3.1](https://github.com/justpranavrs/tlru/compare/v0.3.0...v0.3.1)
- [0.3.0] : [View changes from 0.2.0 to 0.3.0](https://github.com/justpranavrs/tlru/compare/v0.2.0...v0.3.0)
- [0.2.0] : [View changes from 0.1.1 to 0.2.0](https://github.com/justpranavrs/tlru/compare/v0.1.1...v0.2.0)
- [0.1.1] : [View changes from 0.1.0 to 0.1.1](https://github.com/justpranavrs/tlru/compare/v0.1.0...v0.1.1)
- [0.1.0] : [View Initial release commit](https://github.com/justpranavrs/tlru/commit/ccf14bb09d799a6ef439993472e5248655c07a90)
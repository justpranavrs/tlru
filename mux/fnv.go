// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

// FNV Prime for s = 5.
const fnvPrime32 uint32 = 16777619

// NewF32 returns a [Mux] which uses the FNV-1a hash algorithm with
// a custom offset.
//
// Refer, https://www.ietf.org/archive/id/draft-eastlake-fnv-22.html
func NewF32[K comparable](num int) (Mux[K], error) {
	offset := setSeed() // offset is the FNV offset value.
	// It is randomly generated instead of the given FNV offset value
	// to ensure attackers don't brute force keys (Hash DOS) to force the
	// Mux to route all to the same shard.

	mux := getFnvMux[K](offset)
	if mux == nil {
		return *new(Mux[K]), ErrInvalidMuxF32
	}
	return func(key K) uint32 {
		hash := mux(key)
		return fastrange(hash, num)
	}, nil
}

// Get returns the shard number of the corresponding key.
func getFnvMux[K comparable](offset uint32) Mux[K] {
	var mux any

	switch any(*new(K)).(type) {
	case string:
		mux = fnvString(offset)
	case bool:
		mux = fnvBool()
	// int
	case int:
		mux = fnvNumber[int](offset, 8)
	case int8:
		mux = fnvNumber[int8](offset, 1)
	case int16:
		mux = fnvNumber[int16](offset, 2)
	case int32:
		mux = fnvNumber[int32](offset, 4)
	case int64:
		mux = fnvNumber[int64](offset, 8)

	// uint
	case uint:
		mux = fnvNumber[uint](offset, 8)
	case uint8:
		mux = fnvNumber[uint8](offset, 1)
	case uint16:
		mux = fnvNumber[uint16](offset, 2)
	case uint32:
		mux = fnvNumber[uint32](offset, 4)
	case uint64:
		mux = fnvNumber[uint64](offset, 8)
	case uintptr:
		mux = fnvNumber[uintptr](offset, 8)

	default:
		return nil
	}

	if fun, ok := mux.(Mux[K]); ok {
		return fun
	}
	return nil
}

// fnvString returns a [Mux] implements FNV-1 for an input string.
func fnvString(offset uint32) Mux[string] {
	return func(s string) uint32 {
		hash := offset
		for i := 0; i < len(s); i++ {
			hash ^= uint32(s[i])
			hash *= fnvPrime32
		}
		return hash
	}
}

// fnvNumber returns a [Mux] which implements FNV-1a for all numbers.
func fnvNumber[K muxNumber](offset uint32, size int) Mux[K] {
	return func(num K) uint32 {
		hash := offset
		key := uint64(num)

		for i := 0; i < size; i++ {
			hash ^= uint32(key & 255)
			key >>= 8
			hash *= fnvPrime32
		}
		return hash
	}
}

// fnvBool returns a [Mux] which implements FNV-1a for booleans.
func fnvBool() Mux[bool] {
	return func(b bool) uint32 {
		if b {
			return 1
		}
		return 0
	}
}

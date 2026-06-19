// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"math/bits"

	"github.com/justpranavrs/tlru/internal/errs"
)

// xxHash32 Primes
const xxHashPrime1 uint32 = 2654435761
const xxHashPrime2 uint32 = 2246822519
const xxHashPrime3 uint32 = 3266489917
const xxHashPrime4 uint32 = 668265263
const xxHashPrime5 uint32 = 374761393

// NewX32 returns a [Mux] which uses the xxHash32 algorithm
// with a randomly generated xxHash32 seed value
// and initializes the accumulators based on the value.
func NewX32[K comparable](shards int) (Mux[K], error) {
	seed := setSeed() // seed is used to get the initial configuration
	// of the accumulators in xxHash32.

	// acc[1-4] are the accumulators in the xxHash32 algorithm.
	// They are usually represented as v[1-4].
	acc1 := seed + xxHashPrime1 + xxHashPrime2
	acc2 := seed + xxHashPrime2
	acc3 := seed
	acc4 := seed - xxHashPrime1

	mux := getXXHMux[K](seed, acc1, acc2, acc3, acc4)
	if mux == nil {
		return *new(Mux[K]), errs.ErrInvalidMuxX32
	}
	return func(key K) uint32 {
		hash := mux(key)
		return fastrange(hash, shards) // fastrange is better than modulo operator
	}, nil
}

// getXXHMux returns the internal mux function for
// xxHash32 based on the generic key type.
//
// It is used internally in the [NewX32] constructor.
func getXXHMux[K comparable](seed, acc1, acc2, acc3, acc4 uint32) Mux[K] {
	var mux any

	switch any(*new(K)).(type) {
	case string:
		mux = xxHString(seed, acc1, acc2, acc3, acc4)
	case bool:
		mux = xxHBool()

	// int
	case int:
		mux = xxHNumber[int](seed, 8)
	case int8:
		mux = xxHNumber[int8](seed, 1)
	case int16:
		mux = xxHNumber[int16](seed, 2)
	case int32:
		mux = xxHNumber[int32](seed, 4)
	case int64:
		mux = xxHNumber[int64](seed, 8)

		// uint
	case uint:
		mux = xxHNumber[uint](seed, 8)
	case uint8:
		mux = xxHNumber[uint8](seed, 1)
	case uint16:
		mux = xxHNumber[uint16](seed, 2)
	case uint32:
		mux = xxHNumber[uint32](seed, 4)
	case uint64:
		mux = xxHNumber[uint64](seed, 8)
	case uintptr:
		mux = xxHNumber[uintptr](seed, 8)
	default:
		return nil
	}

	if fun, ok := mux.(Mux[K]); ok {
		return fun
	}
	return nil
}

// xxHString returns a [Mux] which implements xxHash32 for a
// string s and returns the output hash.
func xxHString(seed, acc1, acc2, acc3, acc4 uint32) Mux[string] {
	return func(s string) uint32 {
		// process 16 byte blocks, except the final block.
		i := 0
		acc1 := acc1
		acc2 := acc2
		acc3 := acc3
		acc4 := acc4
		for i < len(s)-(len(s)&15) {
			acc1 = accumulate(acc1, s, i)
			i += 4
			acc2 = accumulate(acc2, s, i)
			i += 4
			acc3 = accumulate(acc3, s, i)
			i += 4
			acc4 = accumulate(acc4, s, i)
			i += 4
		}

		// compute the hash
		var hash uint32
		if len(s) < 16 {
			hash = seed + xxHashPrime5
		} else {
			hash = bits.RotateLeft32(acc1, 1) + bits.RotateLeft32(acc2, 7)
			hash += bits.RotateLeft32(acc3, 12) + bits.RotateLeft32(acc4, 18)
		}
		hash += uint32(len(s))

		// process remaining possible 4 byte blocks
		for i < len(s)-(len(s)&3) {
			hash = bits.RotateLeft32((hash+extractFourBytes(s, i)*xxHashPrime3), 17) * xxHashPrime4
			i += 4
		}
		for i < len(s) {
			hash = bits.RotateLeft32((hash+uint32(s[i])*xxHashPrime5), 11) * xxHashPrime1
			i++
		}
		return xxHFinal(hash)
	}
}

// xxHNumber returns a [Mux] which implements xxHash32
// for a number num and returns the output hash.
func xxHNumber[K muxNumber](seed, size uint32) Mux[K] {
	return func(num K) uint32 {
		hash := seed + xxHashPrime5 + size

		key := uint64(num)
		hash = bits.RotateLeft32((hash+(uint32(key)*xxHashPrime3)), 17) * xxHashPrime4
		if size > 4 {
			hash = bits.RotateLeft32((hash+uint32(key>>32)*xxHashPrime3), 17) * xxHashPrime4
		}
		return xxHFinal(hash)
	}
}

// xxHBool returns a [Mux] which implements xxHash32
// for a boolean b and returns the output hash.
func xxHBool() Mux[bool] {
	return func(b bool) uint32 {
		if b {
			return 1
		}
		return 0
	}
}

// xxHFinal performs the final step of xxHash32.
func xxHFinal(hash uint32) uint32 {
	hash ^= (hash >> 15)
	hash *= xxHashPrime2
	hash ^= (hash >> 13)
	hash *= xxHashPrime3
	hash ^= (hash >> 16)

	return hash
}

// accumulate modifies the acc according to the xxHash32 algorithm for a string input.
func accumulate(acc uint32, s string, idx int) uint32 {
	return bits.RotateLeft32((acc+(extractFourBytes(s, idx)*xxHashPrime2)), 13) * xxHashPrime1
}

// extractFourBytes gets the four-byte block from string from idx
func extractFourBytes(s string, idx int) uint32 {
	_ = s[idx+3]
	return uint32(s[idx]) | uint32(s[idx+1])<<8 | uint32(s[idx+2])<<16 | uint32(s[idx+3])<<24
}

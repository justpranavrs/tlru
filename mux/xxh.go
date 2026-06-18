// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"math/bits"
)

// xxHash32 Primes
const xxHashPrime1 uint32 = 2654435761
const xxHashPrime2 uint32 = 2246822519
const xxHashPrime3 uint32 = 3266489917
const xxHashPrime4 uint32 = 668265263
const xxHashPrime5 uint32 = 374761393

// MuxX32 is utilized to route the incoming keys to their correct
// shard in the LRU Cache. It uses the xxHash32 algorithm.
type MuxX32[K comparable] struct {
	// seed is used to get the initial configuration
	// of the accumulators in xxHash32.
	seed uint32

	// acc[1-4] are the accumulators in the xxHash32 algorithm.
	// They are usually represented as v[1-4].
	acc1 uint32
	acc2 uint32
	acc3 uint32
	acc4 uint32

	// mask is used to route the xxHash32 output to its correct shard
	// using bitwise &.
	mask uint32
}

// NewX32 creates a [MuxX32] instance with a randomly generated
// xxHash32 seed value and initializes the accumulators based on the value.
//
// num must be a power of 2 or the Mux would not behave as intended.
func NewX32[K comparable](num int) MuxX32[K] {
	seed := setSeed() // seed for xxHash32
	return MuxX32[K]{
		seed: seed,

		// initializing the accumulators
		acc1: seed + xxHashPrime1 + xxHashPrime2,
		acc2: seed + xxHashPrime2,
		acc3: seed,
		acc4: seed - xxHashPrime1,

		mask: uint32(num) - 1,
	}
}

// Get returns the shard number of the corresponding key.
func (m *MuxX32[K]) Get(key K) (uint32, bool) {
	var hash uint32
	switch t := any(key).(type) {
	case string:
		hash = m.xxHString(t)
	case bool:
		if t {
			return 1 & m.mask, true
		}
		return 0, true

	// int
	case int:
		hash = m.xxHNumber(uint64(t))
	case int8:
		hash = m.xxHNumber(uint64(t))
	case int16:
		hash = m.xxHNumber(uint64(t))
	case int32:
		hash = m.xxHNumber(uint64(t))
	case int64:
		hash = m.xxHNumber(uint64(t))

	// uint
	case uint:
		hash = m.xxHNumber(uint64(t))
	case uint8:
		hash = m.xxHNumber(uint64(t))
	case uint16:
		hash = m.xxHNumber(uint64(t))
	case uint32:
		hash = m.xxHNumber(uint64(t))
	case uint64:
		hash = m.xxHNumber(t)
	case uintptr:
		hash = m.xxHNumber(uint64(t))

	case float32:
		hash = m.xxHNumber(uint64(math.Float32bits(t)))
	case float64:
		hash = m.xxHNumber(math.Float64bits(t))

	default:
		b, err := json.Marshal(key)
		if err != nil {
			return 0, false
		}
		hash = m.xxH(b)
	}

	hash ^= (hash >> 15)
	hash *= xxHashPrime2
	hash ^= (hash >> 13)
	hash *= xxHashPrime3
	hash ^= (hash >> 16)

	return (hash & m.mask), true
}

// xxH implements xxHash for a byte array and returns the output hash.
func (m *MuxX32[K]) xxH(b []byte) uint32 {
	acc1 := m.acc1
	acc2 := m.acc2
	acc3 := m.acc3
	acc4 := m.acc4

	// process 16 byte blocks, except the final block.
	i := 0
	for i < len(b)-(len(b)&15) {
		acc1 = m.accumulate(acc1, b, i)
		i += 4
		acc2 = m.accumulate(acc2, b, i)
		i += 4
		acc3 = m.accumulate(acc3, b, i)
		i += 4
		acc4 = m.accumulate(acc4, b, i)
		i += 4
	}

	// compute the hash
	var hash uint32
	if len(b) < 16 {
		hash = m.seed + xxHashPrime5
	} else {
		hash = bits.RotateLeft32(acc1, 1) + bits.RotateLeft32(acc2, 7)
		hash += bits.RotateLeft32(acc3, 12) + bits.RotateLeft32(acc4, 18)
	}
	hash += uint32(len(b))

	// process remaining possible 4 byte blocks
	for i < len(b)-(len(b)&3) {
		hash = bits.RotateLeft32((hash+m.extractFourBytes(b, i)*xxHashPrime3), 17) * xxHashPrime4
		i += 4
	}
	for i < len(b) {
		hash = bits.RotateLeft32((hash+uint32(b[i])*xxHashPrime5), 11) * xxHashPrime1
		i++
	}
	return hash
}

// xxHString implements xxHash32 for a string s and returns the output hash.
func (m *MuxX32[K]) xxHString(s string) uint32 {
	acc1 := m.acc1
	acc2 := m.acc2
	acc3 := m.acc3
	acc4 := m.acc4

	// process 16 byte blocks, except the final block.
	i := 0
	for i < len(s)-(len(s)&15) {
		acc1 = m.strAccumulate(acc1, s, i)
		i += 4
		acc2 = m.strAccumulate(acc2, s, i)
		i += 4
		acc3 = m.strAccumulate(acc3, s, i)
		i += 4
		acc4 = m.strAccumulate(acc4, s, i)
		i += 4
	}

	// compute the hash
	var hash uint32
	if len(s) < 16 {
		hash = m.seed + xxHashPrime5
	} else {
		hash = bits.RotateLeft32(acc1, 1) + bits.RotateLeft32(acc2, 7)
		hash += bits.RotateLeft32(acc3, 12) + bits.RotateLeft32(acc4, 18)
	}
	hash += uint32(len(s))

	// process remaining possible 4 byte blocks
	for i < len(s)-(len(s)&3) {
		hash = bits.RotateLeft32((hash+m.strExtractFourBytes(s, i)*xxHashPrime3), 17) * xxHashPrime4
		i += 4
	}
	for i < len(s) {
		hash = bits.RotateLeft32((hash+uint32(s[i])*xxHashPrime5), 11) * xxHashPrime1
		i++
	}
	return hash
}

// xxHNumber implements xxHash32 for a number num and returns the output hash.
func (m *MuxX32[K]) xxHNumber(num uint64) uint32 {
	hash := m.seed + xxHashPrime5 + 8

	hash = bits.RotateLeft32((hash+(uint32(num)*xxHashPrime3)), 17) * xxHashPrime4
	hash = bits.RotateLeft32((hash+(uint32(num>>32)*xxHashPrime3)), 17) * xxHashPrime4
	return hash
}

// accumulate modifies the acc according to the xxHash32 algorithm.
func (m *MuxX32[K]) accumulate(acc uint32, b []byte, idx int) uint32 {
	return bits.RotateLeft32((acc+(m.extractFourBytes(b, idx)*xxHashPrime2)), 13) * xxHashPrime1
}

// strAccumulate modifies the acc according to the xxHash32 algorithm for a string input.
func (m *MuxX32[K]) strAccumulate(acc uint32, s string, idx int) uint32 {
	return bits.RotateLeft32((acc+(m.strExtractFourBytes(s, idx)*xxHashPrime2)), 13) * xxHashPrime1
}

// extractFourBytes gets the four-byte block from string from idx
func (m *MuxX32[K]) extractFourBytes(b []byte, idx int) uint32 {
	return binary.LittleEndian.Uint32(b[idx : idx+4])
}

// strExtractFourBytes gets the four-byte block from string from idx
func (m *MuxX32[K]) strExtractFourBytes(s string, idx int) uint32 {
	_ = s[idx+3]
	return uint32(s[idx]) | uint32(s[idx+1])<<8 | uint32(s[idx+2])<<16 | uint32(s[idx+3])<<24
}

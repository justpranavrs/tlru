// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"encoding/json"
	"math"
)

// FNV Prime for s = 5.
const fnvPrime32 uint32 = 16777619

// MuxF32 is utilized to route the incoming keys to their correct
// shard in the LRU Cache. It uses the FNV-1a hash algorithm
type MuxF32[K comparable] struct {
	// offset is the FNV offset value.
	// It is randomly generated instead of the given FNV offset value
	// to ensure attackers don't brute force keys (Hash DOS) to force the
	// MuxF32 to route all to the same shard.
	offset uint32

	// mask is used to route the fnv hash to its correct shard
	// using bitwise &.
	mask uint32
}

// NewF32 creates a [MuxF32] instance with a randomly generated
// FNV offset value and the mask set to (number of shards - 1).
//
// num must be a power of 2 or the Mux would not behave as intended.
func NewF32[K comparable](num int) MuxF32[K] {
	return MuxF32[K]{
		offset: setSeed(),
		mask:   uint32(num) - 1,
	}
}

// Get returns the shard number of the corresponding key.
func (m *MuxF32[K]) Get(key K) (uint32, bool) {
	switch t := any(key).(type) {
	case MuxHash:
		return t.Hash(), true
	case string:
		return m.fnvString(t)
	case bool:
		if t {
			return 1 & m.mask, true
		}
		return 0, true

	// int
	case int:
		return m.fnvNumber(uint64(t), 8)
	case int8:
		return m.fnvNumber(uint64(t), 1)
	case int16:
		return m.fnvNumber(uint64(t), 2)
	case int32:
		return m.fnvNumber(uint64(t), 4)
	case int64:
		return m.fnvNumber(uint64(t), 8)

	// uint
	case uint:
		return m.fnvNumber(uint64(t), 8)
	case uint8:
		return m.fnvNumber(uint64(t), 1)
	case uint16:
		return m.fnvNumber(uint64(t), 2)
	case uint32:
		return m.fnvNumber(uint64(t), 4)
	case uint64:
		return m.fnvNumber(t, 8)
	case uintptr:
		return m.fnvNumber(uint64(t), 8)

	case float32:
		return m.fnvNumber(uint64(math.Float32bits(t)), 4)
	case float64:
		return m.fnvNumber(math.Float64bits(t), 8)

	default:
		buf, err := json.Marshal(key)
		if err != nil {
			return 0, false
		}
		return m.fnv(buf)
	}
}

// fnv implements the Fowler-Noll-Vo hash algorithm for size s = 5.
// Refer, https://www.ietf.org/archive/id/draft-eastlake-fnv-22.html
func (m *MuxF32[K]) fnv(buf []byte) (uint32, bool) {
	hash := m.offset
	for _, b := range buf {
		hash ^= uint32(b)
		hash *= fnvPrime32
	}
	return (hash & m.mask), true
}

// fnvString implements FNV-1a, and takes string as its input.
func (m *MuxF32[K]) fnvString(s string) (uint32, bool) {
	hash := m.offset
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= fnvPrime32
	}
	return (hash & m.mask), true
}

// fnvNumber implements FNV-1a for uint6, must convert number to uint64
// before using this function.
func (m *MuxF32[K]) fnvNumber(num uint64, size int) (uint32, bool) {
	hash := m.offset
	for i := 0; i < size; i++ {
		hash ^= uint32(num & 255)
		num >>= 8
		hash *= fnvPrime32
	}
	return (hash & m.mask), true
}

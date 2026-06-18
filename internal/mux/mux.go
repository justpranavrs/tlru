// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"unsafe"

	"github.com/justpranavrs/tlru/internal/conv"
)

// FNV Prime for s = 5.
const fnvPrime32 uint32 = 16777619

// Mux32 is utilized to route the incoming keys to their correct
// shard in the LRU Cache.
type Mux32[K comparable] struct {
	// offset is the FNV offset value.
	// It is randomly generated instead of the given FNV offset value
	// to ensure attackers don't brute force keys (Hash DOS) to force the
	// Mux32 to route all to the same shard.
	offset uint32

	// mask is used to route the fnv hash to its correct shard
	// using bitwise &.
	mask uint32

	// unsafe is the determination of speed of Mux32.
	//
	// If true, it will do very fast pointer type unsafe conversions for fnv
	// to achieve high routing speeds.
	//
	// If false, it will perform default conversion with guaranteed type safety.
	unsafe bool
}

// New32 creates a [Mux32] instance with a randomly generated
// FNV offset value and the mask set to (number of shards - 1).
//
// num must be a power of 2 or the Mux would not behave as intended.
func New32[K comparable](num int, unsafe bool) Mux32[K] {
	return Mux32[K]{
		offset: setOffset(),
		mask:   uint32(num) - 1,
		unsafe: unsafe,
	}
}

// Get returns the shard number of the corresponding key.
func (m *Mux32[K]) Get(key K) (uint32, bool) {
	if m.unsafe { // faster unsafe pointer based conversions
		switch t := any(key).(type) {
		case string:
			return m.fnv(unsafe.Slice(unsafe.StringData(t), len(t)))
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
			uintptr, float32, float64, bool:

			return m.fnv(unsafe.Slice((*byte)(unsafe.Pointer(&key)), unsafe.Sizeof(key)))
		default:
			buf, err := json.Marshal(key)
			if err != nil {
				return 0, false
			}
			return m.fnv(buf)
		}
	} else { // type safe conversions
		buf, err := conv.ConvBytes(key)
		if err != nil {
			return 0, false
		}
		return m.fnv(buf)
	}
}

// fnv implements the Fowler-Noll-Vo hash algorithm for size s = 5.
// Refer, https://www.ietf.org/archive/id/draft-eastlake-fnv-22.html
func (m *Mux32[K]) fnv(buf []byte) (uint32, bool) {
	hash := m.offset
	for _, b := range buf {
		hash ^= uint32(b)
		hash *= fnvPrime32
	}
	return (hash & m.mask), true
}

// setOffset generates a random 32 bit number for the FNV offset value.
// If an error occurs, it defaults to the recommended offset value.
func setOffset() uint32 {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return 2166136261
	}
	return binary.LittleEndian.Uint32(b)
}

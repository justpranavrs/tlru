// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
)

// Mux is a function that takes in key K of type comparable and output
// a hash of type [uint32].
type Mux[K comparable] func(K) uint32

// muxNumber consists of all primitive number types and all types derived from it.
type muxNumber interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		uintptr
}

var (
	// ErrInvalidMuxF32 is returned by [NewF32] when the key type is invalid for MuxF32.
	ErrInvalidMuxF32 = errors.New("invalid key type for MuxF32: can be only string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr")

	// ErrInvalidMuxX32 is returned by [NewX32] when the key type is invalid for MuxX32.
	ErrInvalidMuxX32 = errors.New("invalid key type for MuxX32: can be only string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr")
)

// fastrange applies quick math instead of modulo to route
// without overflowing index bounds.
func fastrange(hash uint32, shards int) uint32 {
	return uint32(uint64(hash) * uint64(shards) >> 32)
}

// setSeed generates a random 32 bit number.
func setSeed() uint32 {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return 2166136261
	}
	return binary.LittleEndian.Uint32(b)
}

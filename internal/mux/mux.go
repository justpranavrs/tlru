// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"crypto/rand"
	"encoding/binary"
)

// Mux32 is a router for the incoming keys. It routes to the
// correct shard in the Cache.
type Mux32[K comparable] interface {
	// Get returns the shard number of the corresponding key.
	// It returns false if it couldn't parse a JSON key.
	Get(key K) (uint32, bool)
}

// MuxHash allows the Mux to do a custom hash if a Hash method
// is provided on the key of comparable type.
type MuxHash interface {
	// Hash results in a uint32 which is the corresponding shard for a key.
	Hash() uint32
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

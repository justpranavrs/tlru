// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"crypto/rand"
	"encoding/binary"
)

// MuxHash is a function that takes in key K of type comparable and output
// a hash of type [uint32]. It returns false, if it could not output a hash.
type MuxHash[K comparable] func(key K) (uint32, bool)

// setSeed generates a random 32 bit number.
func setSeed() uint32 {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return 2166136261
	}
	return binary.LittleEndian.Uint32(b)
}

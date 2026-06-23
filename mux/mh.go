// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import "hash/maphash"

// NewMH32 returns a [Mux] which uses the "hash/maphash"
// standard package. It is compatible with all key type of type comparable.
//
// It takes in number of shards as its input.
func NewMH32[K comparable](shards int) Mux[K] {
	seed := maphash.MakeSeed()
	return func(key K) uint32 {
		return fastrange(uint32(maphash.Comparable(seed, key)), shards)
	}
}

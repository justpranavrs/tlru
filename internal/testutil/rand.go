// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testutil

import (
	"fmt"
	"math/rand/v2"
)

// User defines the cache test value
type User struct {
	Name  string
	Email string
}

// GenerateZipfData creates an array of CacheOp.
// uses Zipf's distribution to simulate real data.
func GenerateZipfData(keys int, numOps int) []CacheOp {
	ops := make([]CacheOp, numOps)

	rng := rand.New(rand.NewPCG(18, 5))
	zipF := rand.NewZipf(rng, 1.05, 1, uint64(keys))

	for i := range ops {
		key := int(zipF.Uint64())
		name := fmt.Sprintf("user_%d", key)
		var method int
		if i&1 == 0 {
			method = opGet
		} else {
			method = opPut
		}
		ops[i] = CacheOp{
			method: method,
			key:    key,
			value: User{
				Name:  name,
				Email: name + "@gmail.com",
			},
		}
	}

	rand.Shuffle(numOps, func(i, j int) {
		ops[i], ops[j] = ops[j], ops[i]
	})
	return ops
}

// GenerateRandomData creates an array of CacheOp.
// using pseudo random number generators
func GenerateRandomData(keys int, numOps int) []CacheOp {
	ops := make([]CacheOp, numOps)

	for i := range ops {
		key := rand.IntN(keys)
		name := fmt.Sprintf("user_%d", key)
		var method int
		if i&1 == 0 {
			method = opGet
		} else {
			method = opPut
		}
		ops[i] = CacheOp{
			method: method,
			key:    key,
			value: User{
				Name:  name,
				Email: name + "@gmail.com",
			},
		}
	}

	rand.Shuffle(numOps, func(i, j int) {
		ops[i], ops[j] = ops[j], ops[i]
	})
	return ops
}

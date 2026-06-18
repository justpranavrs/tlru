// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mathutil

// NextPower of 2 uses bit smearing to return the next higher power of 2.
func NextPowerOf2(x int) int {
	if x <= 0 {
		return 0
	} else if x >= 1<<30 {
		return 1 << 30
	}

	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return x + 1
}

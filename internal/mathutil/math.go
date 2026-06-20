// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mathutil

import "math/bits"

// NextPower of 2 uses bit smearing to return the next higher power of 2.
func NextPowerOf2(x uint) int {
	if x == 0 {
		return 0
	}
	return 1 << bits.Len(x-1)
}

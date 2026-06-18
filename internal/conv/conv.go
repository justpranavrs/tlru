// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package conv

import (
	"encoding/json"
	"math"
)

// ConvBytes converts the key into an array of bytes
// useful for the FNV-1a hashing algorithm.
func ConvBytes(K any) ([]byte, error) {
	switch t := K.(type) {
	case string:
		return []byte(t), nil

	case bool:
		if t {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	case int:
		return convIntBytes(uint64(t))
	case int8:
		return convIntBytes(uint64(t))
	case int16:
		return convIntBytes(uint64(t))
	case int32:
		return convIntBytes(uint64(t))
	case int64:
		return convIntBytes(uint64(t))

	case uint:
		return convIntBytes(uint64(t))
	case uint8:
		return convIntBytes(uint64(t))
	case uint16:
		return convIntBytes(uint64(t))
	case uint32:
		return convIntBytes(uint64(t))
	case uint64:
		return convIntBytes(t)

	case float32:
		return convIntBytes(uint64(math.Float32bits(t)))
	case float64:
		return convIntBytes(math.Float64bits(t))

	default:
		buf, err := json.Marshal(t)
		if err != nil {
			return nil, err
		}
		return buf, nil
	}
}

// convIntBytes converts a uint64 to an array of bytes
// using classic logical right shifts.
func convIntBytes(x uint64) ([]byte, error) {
	buf := make([]byte, 8)

	buf[0] = byte(x >> 56)
	buf[1] = byte(x >> 48)
	buf[2] = byte(x >> 40)
	buf[3] = byte(x >> 32)
	buf[4] = byte(x >> 24)
	buf[5] = byte(x >> 16)
	buf[6] = byte(x >> 8)
	buf[7] = byte(x)
	return buf, nil
}

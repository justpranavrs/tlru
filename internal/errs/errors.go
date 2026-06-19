// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errs

import "errors"

var (
	// ErrCoreInvalidCapacity is returned by New when the maximum cache capacity is not in [2, 2147483646].
	ErrCoreInvalidCapacity = errors.New("invalid LRU cache capacity: must be in the range [2, 2147483646]")

	// ErrInvalidCapacity is returned by New when the maximum cache capacity is not in [int32] range or greater than the number of shards.
	ErrInvalidCapacity = errors.New("invalid LRU cache capacity: must be in the range of int32 and greater than the number of shards")

	// ErrInvalidShards is returned by New when the number of shards exceeds [int32] range or equals zero.
	ErrInvalidShards = errors.New("invalid number of shards: must be in the range of int32 and greater than 0")
)

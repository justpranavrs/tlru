// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errs

import "errors"

var (
	// ErrCoreInvalidCapacity is returned by New when the maximum cache capacity is less than 2.
	ErrCoreInvalidCapacity = errors.New("invalid LRU cache capacity: must be in the range of uint32 and greater than 1")

	// ErrInvalidCapacity is returned by New when the maximum cache capacity is not greater than the number of shards.
	ErrInvalidCapacity = errors.New("invalid LRU cache capacity: must be in the range of uint32 and greater than the number of shards")

	// ErrInvalidShards is returned by New when the number of shards exceeds [uint32] range.
	ErrInvalidShards = errors.New("invalid number of shards: must be in the range of uint32 and greater than 0")
)

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errs

import "errors"

var (
	// ErrCoreInvalidCapacity is returned by New when the maximum cache capacity is less than 2.
	ErrCoreInvalidCapacity = errors.New("invalid LRU cache capacity: must be greater than 1")

	// ErrInvalidCapacity is returned by New when the maximum cache capacity is not greater than the number of shards.
	ErrInvalidCapacity = errors.New("invalid LRU cache capacity: must be greater than the number of shards")

	// ErrNoShards is returned by New when the number of shards is less than 1
	ErrNoShards = errors.New("invalid number of shards: must be greater than 0")
)

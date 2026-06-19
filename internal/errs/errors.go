// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errs

import "errors"

var (
	// ErrCoreInvalidCapacity is returned by [lrucore.New] when the maximum cache capacity is not in [2, 2147483646].
	ErrCoreInvalidCapacity = errors.New("invalid LRU cache capacity: must be in the range [2, 2147483646]")

	// ErrInvalidCapacity is returned by [tlru.New] when an invalid cache capacity is passed as argument.
	ErrInvalidCapacity = errors.New("invalid LRU cache capacity: must be in the range of int32 and greater than the number of shards")

	// ErrInvalidShards is returned by [tlru.New] when an invalid number of shards is passed using WithShards.
	ErrInvalidShards = errors.New("invalid number of shards: must be in the range of int32 and greater than 0")

	// ErrInvalidMuxX32 is returned by [mux.NewX32] when the key type is invalid for MuxX32.
	ErrInvalidMuxX32 = errors.New("invalid key type for MuxX32: can be only string, bool, int, int8, int16, int32, uint, uint8, uint16, uint32, uintptr")
)

// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tlru

import (
	"time"

	"github.com/justpranavrs/tlru/lruclock"
	"github.com/justpranavrs/tlru/lrucore"
	"github.com/justpranavrs/tlru/mux"
)

// TLRU is the better implementation of [lrucore.TTLCore]. It creates
// many instances of [lrucore.TTLCore] and works based on [LRU].
// It manages a unified clock for all the separate instances.
type TLRU[K comparable, V any] struct {
	coreCluster[K, V, *lrucore.TTLCore[K, V]]
	clock *lruclock.Clock
}

// tlruConfig represents the configuration of [TLRU]. It should be used with [TLRUOption].
type tlruConfig struct {
	lruConfig
	clock *lruclock.Clock
}

// TLRUOption is used to configure [TLRU] when creating an instance using [NewTTL] constructor.
type TLRUOption interface {
	apply(c *tlruConfig) error
}

// tlruOpt represents [TLRU] only options.
type tlruOpt func(c *tlruConfig) error

// apply is an adapter from [tlruOpt] to [TLRUOption].
func (f tlruOpt) apply(c *tlruConfig) error {
	return f(c)
}

// apply is an adapter from [LRUOption] to [TLRUOption].
func (f LRUOption) apply(c *tlruConfig) error {
	return f(&c.lruConfig)
}

// NewTTL creates a [TLRU] instance with the given capacity, expiresAt and options. It creates
// the required [lrucore.TTLCore] instances, initiates the [mux.Mux] for shard routing.
// It defaults to the Mux with hash/maphash algorithm. Check `tlru/mux` package for alternatives.
//
// Returns [ErrInvalidShards] if shards is not in range [1, 1073741823].
//
// Returns [ErrInvalidCapacity] if capacity is not in the range of int32
// and greater than or equal to twice the number of shards.
func NewTTL[K comparable, V any](capacity int, expiresAt time.Duration, opts ...TLRUOption) (*TLRU[K, V], error) {
	// build the config
	cfg := tlruConfig{
		lruConfig: lruConfig{
			shards: DefaultShards,
			mux:    nil,
		},
		clock: nil,
	}
	for _, opt := range opts { // options
		if opt == nil {
			continue
		}
		if err := opt.apply(&cfg); err != nil {
			return nil, err
		}
	}

	// set the mux hash
	var hash mux.Mux[K]
	if cfg.mux != nil {
		if fun, ok := cfg.mux.(mux.Mux[K]); ok {
			hash = fun
		} else {
			return nil, ErrInvalidMuxKey
		}
	} else {
		hash = mux.NewMH32[K](cfg.shards)
	}

	if cfg.clock == nil {
		cfg.clock = lruclock.New(100 * time.Millisecond)
		_ = cfg.clock.Start()
	}

	createShard := func(cap int) (*lrucore.TTLCore[K, V], error) {
		return lrucore.NewTTL[K, V](cap, expiresAt, lrucore.WithClock(cfg.clock))
	}
	cluster, err := buildCluster(capacity, cfg.shards, hash, createShard)
	if err != nil {
		return nil, err
	}

	return &TLRU[K, V]{
		coreCluster: cluster,
		clock:       cfg.clock,
	}, nil
}

// WithClock allows the usage of a custom clock for [TTLCore].
// It is only initialized if "TTL" is enabled.
//
// NOTE: Using WithClock on [NewTTL] will not start the clock. Use [lruclock.Clock.Start] to
// initiate the timer.
func WithClock(clock *lruclock.Clock) tlruOpt {
	return func(c *tlruConfig) error {
		c.clock = clock
		return nil
	}
}

// Close safely closes the background clock when TTL is enabled on the cache.
func (l *TLRU[K, V]) Close() {
	if l.clock != nil {
		l.clock.Stop()
	}
}

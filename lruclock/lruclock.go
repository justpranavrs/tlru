// Copyright (c) 2026 Pranav R S All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lruclock

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Clock is a background clock created for LRU cache when TTL is enabled.
type Clock struct {
	// Epoch is the Unix timestamp when the clock instance was initialized.
	epoch int64

	// tick measures the current time in terms of (duration).
	tick atomic.Int64

	// duration represents the ticker duration
	duration time.Duration

	// ticker makes sure there is no time drift.
	ticker *time.Ticker

	// running represents whether the clock is ticking.
	running atomic.Bool

	// done is a channel to safely exit the goroutine.
	done chan struct{}

	// once helps perform only one Stop.
	once sync.Once
}

var (
	// ErrClockRunning is returned by [Clock.Start] when it is called more than once.
	ErrClockRunning = errors.New("lruclock: clock is already running")
)

// New creates and initializes a new [Clock] with the specified tick duration.
// It is a One-time use clock.
// It can only [Clock.Start] and [Clock.Stop] exactly once.
func New(d time.Duration) *Clock {
	return &Clock{
		epoch:    time.Now().Unix(),
		ticker:   time.NewTicker(d),
		done:     make(chan struct{}),
		duration: d,
	}
}

// Active returns true if the clock is running else returns false.
func (c *Clock) Active() bool {
	return c.running.Load()
}

// Duration returns the time duration of the clock's ticker.
func (c *Clock) Duration() time.Duration {
	return c.duration
}

// Epoch returns the Unix timestamp when the clock instance was initialized.
func (c *Clock) Epoch() int64 {
	return c.epoch
}

// Since returns the ticks elapsed since t.
func (c *Clock) Since(t int64) int64 {
	return c.tick.Load() - t
}

// Start spawns a background goroutine to update a clock tick.
// It can only start the goroutine once.
func (c *Clock) Start() error {
	if !c.running.CompareAndSwap(false, true) {
		return ErrClockRunning
	}

	c.tick.Store(0)
	go func() {
		for {
			select {
			case <-c.done:
				return
			case <-c.ticker.C:
				c.tick.Add(1)
			}
		}
	}()
	return nil
}

// Stop stops the clock and safely exits the goroutine.
func (c *Clock) Stop() {
	c.once.Do(func() {
		c.running.Store(false)
		c.ticker.Stop()
		close(c.done)
	})
}

// Now returns the current tick count.
func (c *Clock) Now() int64 {
	return c.tick.Load()
}

// Until returns the ticks until t.
func (c *Clock) Until(t int64) int64 {
	return t - c.tick.Load()
}

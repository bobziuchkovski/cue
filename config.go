// Copyright (c) 2016 Bob Ziuchkovski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cue

import (
	"sync"
	"sync/atomic"
)

// cfg holds our global logging config.
var cfg = &atomicConfig{}

func init() {
	cfg.set(newConfig())
}

type atomicConfig struct {
	// The mutex is only used during config updates.  Reads are handled via
	// atomic value loads.
	mu  sync.Mutex
	cfg atomic.Value
}

func (ac *atomicConfig) get() *config {
	return ac.cfg.Load().(*config)
}

func (ac *atomicConfig) set(c *config) {
	ac.cfg.Store(c)
}

func (ac *atomicConfig) lock() {
	ac.mu.Lock()
}

func (ac *atomicConfig) unlock() {
	ac.mu.Unlock()
}

type config struct {
	threshold   Level
	frames      int
	errorFrames int
	registry    registry
}

type registry map[Collector]*entry

type entry struct {
	threshold Level
	degraded  bool
	worker    worker
}

func (e *entry) clone() *entry {
	return &entry{
		threshold: e.threshold,
		degraded:  e.degraded,
		worker:    e.worker,
	}
}

func newConfig() *config {
	return &config{
		threshold:   OFF,
		frames:      1,
		errorFrames: 1,
		registry:    make(registry),
	}
}

// clone duplicates configuration for atomic updates.
func (c *config) clone() *config {
	new := &config{
		threshold:   c.threshold,
		frames:      c.frames,
		errorFrames: c.errorFrames,
		registry:    make(registry),
	}
	for collector, entry := range c.registry {
		new.registry[collector] = entry.clone()
	}
	return new
}

// updateThreshold should only be called on a new, cloned config
func (c *config) updateThreshold() {
	max := OFF
	for _, e := range c.registry {
		if e.threshold > max && !e.degraded {
			max = e.threshold
		}
	}
	c.threshold = max
}

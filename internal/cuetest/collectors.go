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

package cuetest

import (
	"sync"
	"time"

	"github.com/remerge/cue"
)

// CapturingCollector captures events that are sent to its Collect method.
type CapturingCollector struct {
	captured []*cue.Event
	cond     *sync.Cond
	mu       sync.Mutex
}

// NewCapturingCollector returns a new CapturingCollector instance.
func NewCapturingCollector() *CapturingCollector {
	c := &CapturingCollector{}
	c.cond = sync.NewCond(&c.mu)
	return c
}

// Collect captures the input event for later inspection.
func (c *CapturingCollector) Collect(event *cue.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.captured = append(c.captured, event)
	c.cond.Broadcast()
	return nil
}

// Captured returns a slice of captured events.
func (c *CapturingCollector) Captured() []*cue.Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	dup := make([]*cue.Event, len(c.captured))
	for i, event := range c.captured {
		dup[i] = event
	}
	return dup
}

// WaitCaptured waits for count events to be captured.  If count events aren't
// captured within maxWait time, it panics.
func (c *CapturingCollector) WaitCaptured(count int, maxWait time.Duration) {
	finished := make(chan struct{})
	go c.waitAsync(count, finished)

	select {
	case <-finished:
		return
	case <-time.After(maxWait):
		panic("WaitCaptured timed-out waiting for events")
	}
}

func (c *CapturingCollector) waitAsync(count int, finished chan struct{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for len(c.captured) != count {
		c.cond.Wait()
	}
	close(finished)
}

// String returns a string representation of the CapturingCollector.
func (c *CapturingCollector) String() string {
	return "CapturingCollector()"
}

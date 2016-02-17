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
	"fmt"
	"sync"
	"time"
)

type capturingCollector struct {
	captured []*Event
	cond     *sync.Cond
	mu       sync.Mutex
}

func newCapturingCollector() *capturingCollector {
	c := &capturingCollector{}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func (c *capturingCollector) Collect(event *Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.captured = append(c.captured, event)
	c.cond.Broadcast()
	return nil
}

func (c *capturingCollector) Captured() []*Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	dup := make([]*Event, len(c.captured))
	for i, event := range c.captured {
		dup[i] = event
	}
	return dup
}

func (c *capturingCollector) WaitCaptured(count int, maxWait time.Duration) {
	finished := make(chan struct{})
	go c.waitAsync(count, finished)

	select {
	case <-finished:
		return
	case <-time.After(maxWait):
		panic("WaitCaptured timed-out waiting for events")
	}
}

func (c *capturingCollector) waitAsync(count int, finished chan struct{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for len(c.captured) != count {
		c.cond.Wait()
	}
	close(finished)
}

func (c *capturingCollector) String() string {
	return "capturingCollector()"
}

type blockingCollector struct {
	collector Collector
	unblocked chan struct{}
}

func newBlockingCollector(c Collector) *blockingCollector {
	return &blockingCollector{
		collector: c,
		unblocked: make(chan struct{}),
	}
}

func (c *blockingCollector) Unblock() {
	close(c.unblocked)
}

func (c *blockingCollector) Collect(event *Event) error {
	<-c.unblocked
	return c.collector.Collect(event)
}

func (c *blockingCollector) String() string {
	return fmt.Sprintf("blockingCollector(target=%s)", c.collector)
}

type failingCollector struct {
	collector    Collector
	succeedAfter int
	failCount    int
}

func newFailingCollector(c Collector, succeedAfter int) *failingCollector {
	return &failingCollector{
		collector:    c,
		succeedAfter: succeedAfter,
	}
}

func (c *failingCollector) Collect(event *Event) error {
	if c.failCount < c.succeedAfter {
		c.failCount++
		return fmt.Errorf("%d more failures before I pass the event to my collector", c.succeedAfter-c.failCount)
	}
	return c.collector.Collect(event)
}

func (c *failingCollector) String() string {
	return fmt.Sprintf("failingCollector(target=%s)", c.collector)
}

type panickingCollector struct {
	collector    Collector
	succeedAfter int
	panicCount   int
}

func newPanickingCollector(c Collector, succeedAfter int) *panickingCollector {
	return &panickingCollector{
		collector:    c,
		succeedAfter: succeedAfter,
	}
}

func (c *panickingCollector) Collect(event *Event) error {
	if c.panicCount < c.succeedAfter {
		c.panicCount++
		panic(fmt.Sprintf("%d more failures before I pass the event to my collector", c.succeedAfter-c.panicCount))
	}
	return c.collector.Collect(event)
}

func (c *panickingCollector) String() string {
	return fmt.Sprintf("panickingCollector(target=%s)", c.collector)
}

type closingCollector struct {
	cond      *sync.Cond
	mu        sync.Mutex
	collector Collector
	closed    bool
}

func newClosingCollector(c Collector) *closingCollector {
	closing := &closingCollector{
		collector: c,
	}
	closing.cond = sync.NewCond(&closing.mu)
	return closing
}

func (c *closingCollector) Collect(event *Event) error {
	return c.collector.Collect(event)
}

func (c *closingCollector) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	c.cond.Broadcast()
	return nil
}

func (c *closingCollector) Closed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.closed
}

func (c *closingCollector) WaitClosed(maxWait time.Duration) {
	finished := make(chan struct{})
	go c.waitAsync(finished)

	select {
	case <-finished:
		return
	case <-time.After(maxWait):
		panic("WaitClosed timed-out waiting for Close() to be called")
	}
}

func (c *closingCollector) waitAsync(finished chan struct{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for !c.closed {
		c.cond.Wait()
	}
	close(finished)
}

func callWithRecover(fn func()) {
	defer func() {
		recover()
	}()
	fn()
}

func callWithLoggerRecover(fn func(), logger Logger, message string) {
	defer logger.Recover(message)
	fn()
}

func callWithLoggerReportRecovery(fn func(), logger Logger, message string) {
	defer func() {
		cause := recover()
		logger.ReportRecovery(cause, message)
	}()
	fn()
}

func resetCue() {
	err := Close(time.Minute)
	if err != nil {
		panic("Cue failed to reset within a minute")
	}
}

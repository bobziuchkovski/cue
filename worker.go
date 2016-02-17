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
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

var errDrops = errors.New("events dropped due to full buffer")

const (
	// Number of collector.Collect() retries before failing an event.
	sendRetries = 2

	// Maximum time to delay between collector.Collect() attempts for a
	// degraded collector.  The backoff is exponentual up to this limit.
	maxDelay = 5 * time.Minute
)

type worker interface {
	Send(event *Event)
	Terminate(flush bool)
}

func newWorker(c Collector, bufsize int) worker {
	if bufsize == 0 {
		return newSyncWorker(c)
	}
	return newAsyncWorker(c, bufsize)
}

type syncWorker struct {
	mu         sync.Mutex
	collector  Collector
	terminated bool
	drops      uint64
}

func newSyncWorker(c Collector) worker {
	return &syncWorker{
		collector: c,
	}
}

func (w *syncWorker) Send(e *Event) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.terminated {
		w.sendEvent(e)
	}
}

func (w *syncWorker) Terminate(flush bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	closeCollector(w.collector)
	w.terminated = true
}

func (w *syncWorker) sendEvent(event *Event) {
	err := sendWithRetries(w.collector, event, sendRetries)
	if err == nil {
		return
	}
	w.drops++
	handleDegradation(w.collector, err, w.drops)
}

type asyncWorker struct {
	// Drops is accessed via atomic operations.  It's the first field to ensure
	// 64-bit alignment.  See the sync/atomic docs for details.
	drops uint64

	collector Collector
	queue     chan *Event
	terminate chan bool
	finished  chan struct{}
	lastdrops uint64
}

func newAsyncWorker(c Collector, bufsize int) worker {
	w := &asyncWorker{
		collector: c,
		queue:     make(chan *Event, bufsize),
		terminate: make(chan bool, 1),
		finished:  make(chan struct{}),
	}
	go w.run()
	return w
}

func (w *asyncWorker) Send(e *Event) {
	select {
	case w.queue <- e:
		// No-op...event is queued
	default:
		atomic.AddUint64(&w.drops, 1)
	}
}

func (w *asyncWorker) run() {
	for {
		select {
		case event := <-w.queue:
			w.handleDrops()
			if event != nil {
				w.sendEvent(event)
			}
		case flush := <-w.terminate:
			w.cleanup(flush)
			close(w.finished)
			return
		}
	}
}

func (w *asyncWorker) Terminate(flush bool) {
	close(w.queue)
	w.terminate <- flush
	close(w.terminate)
	<-w.finished
}

func (w *asyncWorker) cleanup(flush bool) {
	if flush {
		for event := range w.queue {
			w.sendEvent(event)
		}
	}
	closeCollector(w.collector)
	w.queue = nil
}

func (w *asyncWorker) sendEvent(event *Event) {
	err := sendWithRetries(w.collector, event, sendRetries)
	if err == nil {
		return
	}
	drops := atomic.AddUint64(&w.drops, 1)
	handleDegradation(w.collector, err, drops)
	w.lastdrops = drops
}

func (w *asyncWorker) handleDrops() {
	drops := atomic.LoadUint64(&w.drops)
	if drops > w.lastdrops {
		handleDegradation(w.collector, errDrops, drops)
		w.lastdrops = drops
	}
}

func sendWithRetries(c Collector, event *Event, retries int) error {
	defer recoverCollector(c)
	var collectorErr error
	for attempt := 0; attempt <= retries; attempt++ {
		err := c.Collect(event)
		if err == nil {
			return nil
		}
		if collectorErr == nil {
			collectorErr = err
		}
	}
	return collectorErr
}

func handleDegradation(c Collector, err error, drops uint64) {
	defer recoverCollector(c)
	setDegraded(c, true)
	go internalLogger.WithFields(Fields{
		"drops": drops,
	}).Errorf(err, "Collector has entered a degraded state: %s", c)

	ensureErrorSent(c, err, drops)

	setDegraded(c, false)
	go internalLogger.Warnf("Collector has recovered from a degraded stated: %s", c)
}

func ensureErrorSent(c Collector, err error, drops uint64) {
	startTime := time.Now()
	attempt := 0
	for {
		attempt++
		time.Sleep(backoff(attempt))

		ctx := internalContext.WithFields(Fields{
			"attempts": attempt,
			"drops":    drops,
		})
		event := newEventf(ctx, ERROR, err, "The current collector, %s, has been in a degraded state since %s.  Delivery of this message has been attempted %d times", c, startTime.Format(time.Stamp), attempt)
		if c.Collect(event) == nil {
			return
		}
	}
}

func closeCollector(c Collector) {
	closer, ok := c.(io.Closer)
	if !ok {
		return
	}
	internalLogger.Errorf(closer.Close(), "Failed to close collector %s", c)
}

func recoverCollector(c Collector) {
	cause := recover()
	if cause == nil {
		return
	}

	go func() {
		dispose(c)
		message := fmt.Sprintf("Recovered from collector panic. Collector has been disposed: %s", c)
		internalLogger.ReportRecovery(cause, message)
	}()
}

func backoff(attempt int) time.Duration {
	exp := math.Pow(2, float64(attempt))
	if math.IsNaN(exp) || math.IsInf(exp, 1) || math.IsInf(exp, -1) {
		return maxDelay
	}

	delay := time.Millisecond * time.Duration(exp)
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

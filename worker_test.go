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
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewWorker(t *testing.T) {
	w := newWorker(newCapturingCollector(), 0)
	checkSync(t, w)

	w = newWorker(newCapturingCollector(), 1)
	checkAsync(t, w)
}

func TestSyncWorkerSend(t *testing.T) {
	c := newCapturingCollector()
	w := newWorker(c, 0)
	checkSync(t, w)

	w.Send(&Event{})
	if len(c.Captured()) != 1 {
		t.Errorf("Expected to see 1 event, but saw %d instead", len(c.Captured()))
	}
}

func TestSyncWorkerRetry(t *testing.T) {
	c := newCapturingCollector()
	w := newWorker(newFailingCollector(c, sendRetries), 0)
	checkSync(t, w)

	e := &Event{}
	w.Send(e)
	if c.Captured()[0] != e {
		t.Errorf("Expected to see our event, but but saw %#v instead", c.Captured()[0])
	}
}

func TestSyncWorkerDegredation(t *testing.T) {
	// t.Skip("blah")
	defer resetCue()
	c1 := newCapturingCollector()
	Collect(INFO, c1)

	c2 := newCapturingCollector()
	Collect(DEBUG, newFailingCollector(c2, sendRetries+1))

	log := NewLogger("test")
	log.Debug("message")

	c1.WaitCaptured(2, 5*time.Second)
	c2.WaitCaptured(2, 5*time.Second)

	if len(c1.Captured()) != 2 {
		t.Errorf("Expected to see exactly 2 events sent to c1, but saw %d instead", len(c1.Captured()))
	}
	if c1.Captured()[0].Level != ERROR || !strings.Contains(c1.Captured()[0].Message, "Collector has entered a degraded state") {
		t.Errorf("Expected to see a degradation message sent to c1, but saw %#v instead", c1.Captured()[0])
	}
	if c1.Captured()[1].Level != WARN || !strings.Contains(c1.Captured()[1].Message, "Collector has recovered from a degraded stated") {
		t.Errorf("Expected to see a recovery message sent to c1, but saw %#v instead", c1.Captured()[1])
	}
	if len(c2.Captured()) != 2 {
		t.Errorf("Expected to see exactly 2 events sent to c2, but saw %d instead", len(c2.Captured()))
	}
	if c2.Captured()[0].Level != ERROR || !strings.Contains(c2.Captured()[0].Message, "The current collector") || !strings.Contains(c2.Captured()[0].Message, "has been in a degraded state since") {
		t.Errorf("Expected to see a degredation message sent to c2, but saw %#v instead", c2.Captured()[0])
	}
	if c2.Captured()[1].Level != WARN || !strings.Contains(c2.Captured()[1].Message, "Collector has recovered from a degraded stated") {
		t.Errorf("Expected to see a recovery message sent to c2, but saw %#v instead", c2.Captured()[1])
	}
}

func TestSyncWorkerPanic(t *testing.T) {
	defer resetCue()
	c1 := newCapturingCollector()
	Collect(DEBUG, c1)

	c2 := newCapturingCollector()
	w := newWorker(newPanickingCollector(c2, 1), 0)
	checkSync(t, w)

	e := &Event{}
	w.Send(e)

	c1.WaitCaptured(1, 5*time.Second)
	if c1.Captured()[0].Level != FATAL || !strings.Contains(c1.Captured()[0].Message, "Recovered from collector panic. Collector has been disposed") {
		t.Errorf("Expected to see collector panic sent to c1, but saw %#v instead", c1.Captured()[0])
	}
	if len(c2.Captured()) != 0 {
		t.Errorf("Expected to see 0 events collected by c2, but saw %d instead", len(c2.Captured()))
	}
}

func TestSyncWorkerTerminate(t *testing.T) {
	c := newCapturingCollector()
	closing := newClosingCollector(c)
	w := newWorker(closing, 0)
	checkSync(t, w)

	w.Send(&Event{})
	w.Terminate(true)
	w.Send(&Event{})
	if len(c.Captured()) != 1 {
		t.Errorf("Expected to see 1 event, but saw %d instead", len(c.Captured()))
	}
	closing.WaitClosed(5 * time.Second)
}

func TestAsyncWorkerSend(t *testing.T) {
	c := newCapturingCollector()
	w := newWorker(c, 10)
	checkAsync(t, w)

	w.Send(&Event{})
	w.Terminate(true)
	if len(c.Captured()) != 1 {
		t.Errorf("Expected to see 1 event, but saw %d instead", len(c.Captured()))
	}
}

func TestAsyncWorkerSendQueueFull(t *testing.T) {
	defer resetCue()
	c1 := newCapturingCollector()
	Collect(DEBUG, c1)

	c2 := newCapturingCollector()
	blocking := newBlockingCollector(c2)
	w := newWorker(blocking, 1)
	checkAsync(t, w)

	e1 := &Event{Level: DEBUG, Message: "Original, blocked message"}
	w.Send(e1)

	e2 := &Event{}
	w.Send(e2)

	c1.WaitCaptured(1, 5*time.Second)

	if len(c2.Captured()) != 0 {
		t.Errorf("Expected to see 0 events sent to c2 while blocked, but saw %d instead", len(c2.Captured()))
	}
	if len(c1.Captured()) != 1 {
		t.Errorf("Expected to see exactly 1 events sent to c1 while c2 is blocked, but saw %d instead", len(c1.Captured()))
	}
	if c1.Captured()[0].Level != ERROR || !strings.Contains(c1.Captured()[0].Message, "Collector has entered a degraded state") {
		t.Errorf("Expected to see a degradation message sent to c1, but saw %#v instead", c1.Captured()[0])
	}

	blocking.Unblock()
	c1.WaitCaptured(2, 5*time.Second)
	c2.WaitCaptured(2, 5*time.Second)

	if len(c1.Captured()) != 2 {
		t.Errorf("Expected to see exactly 2 events collected by c1 after c2 is unblocked, but saw %d instead", len(c1.Captured()))
	}
	if c1.Captured()[1].Level != WARN || !strings.Contains(c1.Captured()[1].Message, "Collector has recovered from a degraded stated") {
		t.Errorf("Expected to see a recovery message sent to c1, but saw %#v instead", c1.Captured()[1])
	}

	if len(c2.Captured()) != 2 {
		t.Errorf("Expected to see 2 events sent to c2 after being unblocked, but saw %d instead", len(c2.Captured()))
	}
	if c2.Captured()[0].Level != ERROR || !strings.Contains(c2.Captured()[0].Message, "The current collector") || !strings.Contains(c2.Captured()[0].Message, "has been in a degraded state since") {
		t.Errorf("Expected to see a degredation message sent to c2, but saw %#v instead", c2.Captured()[0])
	}
	if c2.Captured()[1].Level != DEBUG || c2.Captured()[1].Message != "Original, blocked message" {
		t.Errorf("Expected to see the blocked message delivered to c2 after being unblocked, but saw %#v instead", c2.Captured()[1])
	}
}

func TestAsyncWorkerRetry(t *testing.T) {
	c := newCapturingCollector()
	w := newWorker(newFailingCollector(c, sendRetries), 10)
	checkAsync(t, w)

	e := &Event{}
	w.Send(e)
	c.WaitCaptured(1, 5*time.Second)
	if c.Captured()[0] != e {
		t.Errorf("Expected to see our event, but but saw %#v instead", c.Captured()[0])
	}
}

func TestAsyncWorkerDegredation(t *testing.T) {
	// t.Skip("blah")
	defer resetCue()
	c1 := newCapturingCollector()
	Collect(INFO, c1)

	c2 := newCapturingCollector()
	CollectAsync(DEBUG, 10, newFailingCollector(c2, sendRetries+1))

	log := NewLogger("test")
	log.Debug("message")

	c1.WaitCaptured(2, 5*time.Second)
	c2.WaitCaptured(2, 5*time.Second)

	if len(c1.Captured()) != 2 {
		t.Errorf("Expected to see exactly 2 events sent to c1, but saw %d instead", len(c1.Captured()))
	}
	if c1.Captured()[0].Level != ERROR || !strings.Contains(c1.Captured()[0].Message, "Collector has entered a degraded state") {
		t.Errorf("Expected to see a degradation message sent to c1, but saw %#v instead", c1.Captured()[0])
	}
	if c1.Captured()[1].Level != WARN || !strings.Contains(c1.Captured()[1].Message, "Collector has recovered from a degraded stated") {
		t.Errorf("Expected to see a recovery message sent to c1, but saw %#v instead", c1.Captured()[1])
	}
	if len(c2.Captured()) != 2 {
		t.Errorf("Expected to see exactly 2 events sent to c2, but saw %d instead", len(c2.Captured()))
	}
	if c2.Captured()[0].Level != ERROR || !strings.Contains(c2.Captured()[0].Message, "The current collector") || !strings.Contains(c2.Captured()[0].Message, "has been in a degraded state since") {
		t.Errorf("Expected to see a degredation message sent to c2, but saw %#v instead", c2.Captured()[0])
	}
	if c2.Captured()[1].Level != WARN || !strings.Contains(c2.Captured()[1].Message, "Collector has recovered from a degraded stated") {
		t.Errorf("Expected to see a recovery message sent to c2, but saw %#v instead", c2.Captured()[1])
	}
}

func TestAsyncWorkerPanic(t *testing.T) {
	defer resetCue()
	c1 := newCapturingCollector()
	Collect(DEBUG, c1)

	c2 := newCapturingCollector()
	w := newWorker(newPanickingCollector(c2, 1), 10)
	checkAsync(t, w)

	e := &Event{}
	w.Send(e)

	c1.WaitCaptured(1, 5*time.Second)
	if len(c1.Captured()) != 1 {
		t.Errorf("Expected to see exactly 1 events sent to c1, but saw %d instead", len(c1.Captured()))
	}
	if c1.Captured()[0].Level != FATAL || !strings.Contains(c1.Captured()[0].Message, "Recovered from collector panic. Collector has been disposed") {
		t.Errorf("Expected to see collector panic sent to c1, but saw %#v instead", c1.Captured()[0])
	}
	if len(c2.Captured()) != 0 {
		t.Errorf("Expected to see 0 events collected by c2, but saw %d instead", len(c2.Captured()))
	}
}

func TestAsyncWorkerTerminate(t *testing.T) {
	c := newCapturingCollector()
	blocking := newBlockingCollector(c)
	closing := newClosingCollector(blocking)
	w := newWorker(closing, 20)
	checkAsync(t, w)

	w.Send(&Event{})
	w.Send(&Event{})
	go blocking.Unblock()
	w.Terminate(true)

	c.WaitCaptured(2, 5*time.Second)
	if len(c.Captured()) != 2 {
		t.Errorf("Expected to see 2 events, but saw %d instead", len(c.Captured()))
	}
	closing.WaitClosed(5 * time.Second)
}

func TestBackoff(t *testing.T) {
	if backoff(1) < time.Millisecond {
		t.Errorf("Expected a minimum backoff delay of no less than 1 ms, but saw %s instead", backoff(1))
	}
	if backoff(50) > time.Hour {
		t.Errorf("Expected a maximum backoff delay of no more than 1 hr, but saw %s instead", backoff(50))
	}
	// Test handling of exponential overflow
	if backoff(1000000) < time.Second || backoff(1000000) > time.Hour {
		t.Errorf("Expected to see a reasonable delay for a very high attempt, but saw %s instead", backoff(1000000))
	}
}

func checkSync(t *testing.T, worker worker) {
	_, ok := worker.(*syncWorker)
	if !ok {
		t.Errorf("Expected *cue.syncWorker but got %s instead", reflect.TypeOf(worker))
	}
}

func checkAsync(t *testing.T, worker worker) {
	_, ok := worker.(*asyncWorker)
	if !ok {
		t.Errorf("Expected *cue.asyncWorker but got %s instead", reflect.TypeOf(worker))
	}
}

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
	"reflect"
	"strings"
	"testing"
	"time"
)

var loggerContextTests = []struct {
	Name       string
	Logger     Logger
	FieldEquiv Fields
}{
	{
		Name:       "Empty",
		Logger:     NewLogger("Empty"),
		FieldEquiv: Fields{},
	},
	{
		Name:       "WithValue",
		Logger:     NewLogger("WithValue").WithValue("k1", "v1"),
		FieldEquiv: Fields{"k1": "v1"},
	},
	{
		Name:       "WithFields",
		Logger:     NewLogger("WithFields").WithFields(Fields{"k1": "v1", "k2": 2}),
		FieldEquiv: Fields{"k1": "v1", "k2": 2},
	},
	{
		Name:       "Chained1",
		Logger:     NewLogger("Chained1").WithFields(Fields{"k1": "v1", "k2": 2}).WithValue("k3", 3.0),
		FieldEquiv: Fields{"k1": "v1", "k2": 2, "k3": 3.0},
	},
	{
		Name:       "Chained2",
		Logger:     NewLogger("Chained2").WithValue("k1", "v1").WithFields(Fields{"k2": 2, "k3": 3.0}),
		FieldEquiv: Fields{"k1": "v1", "k2": 2, "k3": 3.0},
	},
}

func TestLoggerContext(t *testing.T) {
	for _, test := range loggerContextTests {
		defer resetCue()
		c := newCapturingCollector()
		Collect(DEBUG, c)

		test.Logger.Debug("test")
		event := c.Captured()[0]
		if !reflect.DeepEqual(event.Context.Fields(), test.FieldEquiv) {
			t.Errorf("Logger context is incorrect.  Test: %s, Expected: %v, Received: %v", test.Name, test.FieldEquiv, event.Context.Fields())
		}
	}
}

func TestLoggerDebug(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	log := NewLogger("test")
	log.Debug("Debug Test")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], DEBUG, "Debug Test", nil)
}

func TestLoggerDebugf(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	log := NewLogger("test")
	log.Debugf("Debugf %s", "Test")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], DEBUG, "Debugf Test", nil)
}

func TestLoggerInfo(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	log := NewLogger("test")
	log.Info("Info Test")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], INFO, "Info Test", nil)
}

func TestLoggerInfof(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	log := NewLogger("test")
	log.Infof("Infof %s", "Test")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], INFO, "Infof Test", nil)
}

func TestLoggerWarn(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	log := NewLogger("test")
	log.Warn("Warn Test")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], WARN, "Warn Test", nil)
}

func TestLoggerWarnf(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	log := NewLogger("test")
	log.Warnf("Warnf %s", "Test")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], WARN, "Warnf Test", nil)
}

func TestLoggerError(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	cause := errors.New("Error Cause")
	log := NewLogger("test")
	result := log.Error(cause, "Error Test")
	if result != cause {
		t.Error("Expected to receive the same error cause as the return value but dind't")
	}
	log.Error(nil, "Error Test, nil")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], ERROR, "Error Test", cause)
}

func TestLoggerErrorf(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	cause := errors.New("Errorf Cause")
	log := NewLogger("test")
	result := log.Errorf(cause, "Errorf %s", "Test")
	if result != cause {
		t.Error("Expected to receive the same error cause as the return value but dind't")
	}
	log.Errorf(nil, "Errorf %s, nil", "Test")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], ERROR, "Errorf Test", cause)
}

func TestLoggerPanic(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	cause := errors.New("Panic Cause")
	callWithRecover(func() {
		log := NewLogger("test")
		log.Panic(cause, "Panic Test, error")
	})
	callWithRecover(func() {
		log := NewLogger("test")
		log.Panic("other", "Panic Test, other")
	})
	callWithRecover(func() {
		log := NewLogger("test")
		log.Panic(nil, "Panic Test, nil")
	})

	if len(c.Captured()) != 2 {
		t.Errorf("Expected 2 log events but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], FATAL, "Panic Test, error", cause)

	// We can't use checkEventExpectation with the "other" cause
	if c.Captured()[1].Level != FATAL {
		t.Errorf("Expected to see FATAL level for panic with %q cause, but saw %s instead", "other", c.Captured()[1].Level)
	}
	if c.Captured()[1].Message != "Panic Test, other" {
		t.Errorf("Expected to see %q message for panic with %q cause, but saw %s instead", "Panic Test, other", "other", c.Captured()[1].Message)
	}
	if c.Captured()[1].Error.Error() != "other" {
		t.Errorf("Expected to see %q error for panic with %q cause, but saw %s instead", "other", "other", c.captured[1].Error.Error())
	}
}

func TestLoggerPanicf(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	cause := errors.New("Panicf Cause")
	callWithRecover(func() {
		log := NewLogger("test")
		log.Panicf(cause, "Panicf %s, error", "Test")
	})
	callWithRecover(func() {
		cause := "other"
		log := NewLogger("test")
		log.Panicf(cause, "Panicf %s, other", "Test")
	})
	callWithRecover(func() {
		log := NewLogger("test")
		log.Panicf(nil, "Panicf %s, nil", "Test")
	})

	if len(c.Captured()) != 2 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], FATAL, "Panicf Test, error", cause)

	// We can't use checkEventExpectation with the "other" cause
	if c.Captured()[1].Level != FATAL {
		t.Errorf("Expected to see FATAL level for panicf with %q cause, but saw %s instead", "other", c.Captured()[1].Level)
	}
	if c.Captured()[1].Message != "Panicf Test, other" {
		t.Errorf("Expected to see %q message for panicf with %q cause, but saw %s instead", "Panicf Test, other", "other", c.Captured()[1].Message)
	}
	if c.Captured()[1].Error.Error() != "other" {
		t.Errorf("Expected to see %q error for panicf with %q cause, but saw %s instead", "other", "other", c.captured[1].Error.Error())
	}
}

func TestLoggerRecoverPanic(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	cause := errors.New("Recover Test")
	callWithLoggerRecover(func() {
		panic(cause)
	}, NewLogger("test"), "Recover Test")
	callWithLoggerRecover(func() {
		// No-op
	}, NewLogger("test"), "Recover Test, nil")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], FATAL, "Recover Test", cause)
}

func TestLoggerRecoverPanicMethod(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	cause := errors.New("Recover Panic Method Test")
	log := NewLogger("test")
	callWithLoggerRecover(func() {
		log.Panic(cause, "Panic")
	}, log, "Recover Panic Method Test")
	callWithLoggerRecover(func() {
		log.Panic(nil, "Panic, nil")
	}, log, "Recover Panic Method Test, nil")

	// Since we panic from our own logger, the message should be the log.Panic message
	checkEventExpectation(t, c.Captured()[0], FATAL, "Panic", cause)

	// We should detect the panic method call and not emit two events
	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
}

func TestLoggerRecoverPanicfMethod(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	cause := errors.New("Recover Panicf Method Test")
	log := NewLogger("test")
	callWithLoggerRecover(func() {
		log.Panicf(cause, "Panicf %s", "Test")
	}, log, "Recover Panicf Method Test")
	callWithLoggerRecover(func() {
		log.Panicf(nil, "Panicf %s, nil", "Test")
	}, log, "Recover Panicf Method Test, nil")

	// Since we panic from our own logger, the message should be the log.Panicf message
	checkEventExpectation(t, c.Captured()[0], FATAL, "Panicf Test", cause)

	// We should detect the panic method call and not emit two events
	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
}

func TestLoggerRecoverNoop(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	callWithLoggerRecover(func() {}, NewLogger("test"), "Recover No-op Test")

	// There's no panic, so nothing should be logged
	if len(c.Captured()) != 0 {
		t.Errorf("Expected to receive 0 events but received %d", len(c.Captured()))
	}
}

func TestLoggerReportRecoveryPanic(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	cause := errors.New("Report Recovery Test")
	callWithLoggerReportRecovery(func() {
		panic(cause)
	}, NewLogger("test"), "Report Recovery Test")
	callWithLoggerReportRecovery(func() {
		// No-op
	}, NewLogger("test"), "Report Recovery Test, nil")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
	checkEventExpectation(t, c.Captured()[0], FATAL, "Report Recovery Test", cause)
}

func TestLoggerReportRecoveryPanicMethod(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	cause := errors.New("Report Recovery Panic Method Test")
	log := NewLogger("test")
	callWithLoggerReportRecovery(func() {
		log.Panic(cause, "Panic")
	}, log, "Report Recovery Panic Method Test")
	callWithLoggerReportRecovery(func() {
		log.Panic(nil, "Panic, nil")
	}, log, "Report Recovery Panic Method Test, nil")

	// Since we panic from our own logger, the message should be the log.Panic message
	checkEventExpectation(t, c.Captured()[0], FATAL, "Panic", cause)

	// We should detect the panic method call and not emit two events
	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
}

func TestLoggerReportRecoveryPanicfMethod(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	cause := errors.New("Report Recovery Panicf Method Test")
	log := NewLogger("test")
	callWithLoggerReportRecovery(func() {
		log.Panicf(cause, "Panicf %s", "Test")
	}, log, "Report Recovery Panicf Method Test")
	callWithLoggerReportRecovery(func() {
		log.Panicf(nil, "Panicf %s, nil", "Test")
	}, log, "Report Recovery Panicf Method Test, nil")

	// Since we panic from our own logger, the message should be the log.Panicf message
	checkEventExpectation(t, c.Captured()[0], FATAL, "Panicf Test", cause)

	// We should detect the panic method call and not emit two events
	if len(c.Captured()) != 1 {
		t.Errorf("Expected only a single log event but received %d", len(c.Captured()))
	}
}

func TestLoggerReportRecoveryNoop(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	callWithLoggerReportRecovery(func() {}, NewLogger("test"), "Report Recovery No-op Test")

	// There's no panic, so nothing should be logged
	if len(c.Captured()) != 0 {
		t.Errorf("Expected to receive 0 events but received %d", len(c.Captured()))
	}
}

type wrappedLogger struct {
	Logger Logger
}

func (w wrappedLogger) Info(message string) {
	w.Logger.Info(message)
}

func TestLoggerWrap(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	unwrapped := NewLogger("unwrapped")
	wrapped := unwrapped.Wrap()

	unwrapped.Info("unwrapped logger message from test func")
	wrappedLogger{Logger: unwrapped}.Info("unwrapped logger message from wrapper type")
	wrappedLogger{Logger: wrapped}.Info("wrapped logger message from wrapper type")

	if len(c.Captured()) != 3 {
		t.Errorf("Expected to receive 3 events but received %d", len(c.Captured()))
	}
	source1 := c.Captured()[0].Frames[0]
	source2 := c.Captured()[1].Frames[0]
	source3 := c.Captured()[2].Frames[0]
	thisfunc := "github.com/remerge/cue.TestLoggerWrap"
	if source1.Function != "github.com/remerge/cue.TestLoggerWrap" {
		t.Errorf("First event has incorrect source function.  Expected %s, Received %s", thisfunc, source1.Function)
	}
	if source2.Function == "github.com/remerge/cue.TestLoggerWrap" {
		t.Errorf("Second event has incorrect source function.  Expected it NOT to match %s, but it matches.", thisfunc)
	}
	if source3.Function != "github.com/remerge/cue.TestLoggerWrap" {
		t.Errorf("Third event has incorrect source function.  Expected %s, Received %s", thisfunc, source3.Function)
	}
}

func TestThresholds(t *testing.T) {
	defer resetCue()

	debugc := newCapturingCollector()
	infoc := newCapturingCollector()
	warnc := newCapturingCollector()
	errorc := newCapturingCollector()
	fatalc := newCapturingCollector()
	offc := newCapturingCollector()

	log := NewLogger("test")
	log.Debug("Uncollected Debug event")
	log.Debugf("Uncollected Debugf %s", "event")
	callWithLoggerRecover(func() { panic("Uncollected Panic") }, log, "Uncollected Recover")

	Collect(DEBUG, debugc)
	Collect(INFO, infoc)
	Collect(WARN, warnc)
	Collect(ERROR, errorc)
	Collect(FATAL, fatalc)
	Collect(OFF, offc)

	log.Debug("Debug event")
	log.Debugf("Debugf %s", "event")
	log.Info("Info event")
	log.Infof("Infof %s", "event")
	log.Warn("Warn event")
	log.Warnf("Warnf %s", "event")
	log.Error(errors.New("Error event"), "Error event")
	log.Errorf(errors.New("Errorf event"), "Errorf %s", "event")

	cause := errors.New("panic cause")
	callWithRecover(func() {
		log.Panicf(cause, "Panicf %s", "event")
	})
	callWithRecover(func() {
		log.Panic(cause, "Panic event")
	})

	if len(debugc.Captured()) != 10 {
		t.Errorf("Expected collector at DEBUG threshold to receive 10 events, but it received %d instead", len(debugc.Captured()))
	}
	if len(infoc.Captured()) != 8 {
		t.Errorf("Expected collector at INFO threshold to receive 8 events, but it received %d instead", len(infoc.Captured()))
	}
	if len(warnc.Captured()) != 6 {
		t.Errorf("Expected collector at WARN threshold to receive 6 events, but it received %d instead", len(warnc.Captured()))
	}
	if len(errorc.Captured()) != 4 {
		t.Errorf("Expected collector at ERROR threshold to receive 4 events, but it received %d instead", len(errorc.Captured()))
	}
	if len(fatalc.Captured()) != 2 {
		t.Errorf("Expected collector at FATAL threshold to receive 2 events, but it received %d instead", len(fatalc.Captured()))
	}
	if len(offc.Captured()) != 0 {
		t.Errorf("Expected collector at OFF threshold to receive 0 events, but it received %d instead", len(offc.Captured()))
	}
}

func TestClose(t *testing.T) {
	defer resetCue()
	sync := newCapturingCollector()
	async := newCapturingCollector()
	blocking := newBlockingCollector(async)
	blocking.Unblock()
	Collect(DEBUG, sync)
	CollectAsync(DEBUG, 100, async)

	log := NewLogger("test")
	log.Debug("message 1")

	err := Close(time.Minute)
	if err != nil {
		panic("Failed to close within a minute.  Panicking because we are now in an unknown state.")
	}

	if len(sync.Captured()) != 1 {
		t.Errorf("Expected to collect exactly 1 sync event but received %d instead", len(sync.Captured()))
	}
	if len(async.Captured()) != 1 {
		t.Errorf("Expected to collect exactly 1 async event but received %d instead", len(async.Captured()))
	}

	log.Debug("message 2")
	if len(sync.Captured()) != 1 {
		t.Errorf("Expected to STILL have exactly 1 sync event but now have %d instead", len(sync.Captured()))
	}
	if len(async.Captured()) != 1 {
		t.Errorf("Expected to STILL have exactly 1 async event but now have %d instead", len(async.Captured()))
	}

	Collect(DEBUG, sync)
	CollectAsync(DEBUG, 100, async)
	log.Debug("message 3")

	err = Close(time.Minute)
	if err != nil {
		panic("Failed to close within a minute.  Panicking because we are now in an unknown state.")
	}
	if len(sync.Captured()) != 2 {
		t.Errorf("Expected a total of 2 sync events but now have %d instead", len(sync.Captured()))
	}
	if len(async.Captured()) != 2 {
		t.Errorf("Expected a total of 2 async event but received %d instead", len(async.Captured()))
	}
}

func TestCloseTimeout(t *testing.T) {
	defer resetCue()
	async := newCapturingCollector()
	blocking := newBlockingCollector(async)
	defer blocking.Unblock()
	CollectAsync(DEBUG, 10, blocking)

	log := NewLogger("test")
	log.Debug("message 1")
	log.Debug("message 2")

	err := Close(50 * time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Error("Expected to see timeout error waiting for blocked worker to flush")
	}
}

func TestCloseNoop(t *testing.T) {
	defer resetCue()
	err := Close(time.Minute)
	if err != nil {
		panic("Close should be a no-op when we're in a blank state")
	}
}

func TestCollect(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	log := NewLogger("test")
	log.Debug("message 1")
	if len(c.Captured()) != 1 {
		t.Errorf("Expected to collect exactly 1 event but received %d instead", len(c.Captured()))
	}
}

func TestCollectNilCollector(t *testing.T) {
	// Check to make sure nothing blows up
	defer resetCue()
	Collect(DEBUG, nil)
	log := NewLogger("test")
	log.Debug("message")
}

func TestCollectDuplicateCollector(t *testing.T) {
	// Check to make sure nothing blows up and threshold doesn't change
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)
	Collect(INFO, c)

	log := NewLogger("test")
	log.Debug("message")

	if len(c.Captured()) != 1 {
		t.Errorf("Expected to collect exactly 1 event but received %d instead", len(c.Captured()))
	}
}

func TestCollectAsync(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	blocking := newBlockingCollector(c)
	CollectAsync(DEBUG, 100, blocking)

	log := NewLogger("test")
	for i := 0; i < 100; i++ {
		log.Debugf("message %i")
	}
	captured := c.Captured()
	if len(captured) != 0 {
		t.Errorf("Expected 0 events to be delivered, but %d were delivered instead.", len(captured))
	}

	blocking.Unblock()
	err := Close(time.Minute)
	if err != nil {
		panic("Failed to close within a minute.  Panicking because we are now in an unknown state.")
	}
	captured = c.Captured()
	if len(captured) != 100 {
		t.Errorf("Expected 100 events to be delivered, but %d were delivered instead.", len(captured))
	}
}

func TestDispose(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	log := NewLogger("test")
	log.Debug("message 1")
	dispose(c)
	log.Debug("message 2")
	if len(c.Captured()) != 1 {
		t.Errorf("Expected to collect exactly 1 event but received %d instead", len(c.Captured()))
	}
}

func TestSetFrames(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	log := NewLogger("test")
	SetFrames(0, 0)
	log.Debug("message 1")
	SetFrames(2, 3)
	log.Debug("message 2")
	log.Error(errors.New("test"), "message 3")
	SetFrames(0, 1)
	log.Debug("message 4")
	SetFrames(1, 0)
	log.Error(errors.New("test"), "message 5")
	if len(c.Captured()[0].Frames) != 0 {
		t.Errorf("Expected message 1 to have 0 frames, but it had %d instead", len(c.Captured()[0].Frames))
	}
	if len(c.Captured()[1].Frames) != 2 {
		t.Errorf("Expected message 2 to have 2 frames, but it had %d instead", len(c.Captured()[1].Frames))
	}
	if len(c.Captured()[2].Frames) != 3 {
		t.Errorf("Expected message 3 to have 3 frames, but it had %d instead", len(c.Captured()[2].Frames))
	}
	if len(c.Captured()[3].Frames) != 0 {
		t.Errorf("Expected message 4 to have 0 frames, but it had %d instead", len(c.Captured()[3].Frames))
	}
	if len(c.Captured()[4].Frames) != 0 {
		t.Errorf("Expected message 5 to have 0 frames, but it had %d instead", len(c.Captured()[4].Frames))
	}
}

func TestSetLevel(t *testing.T) {
	defer resetCue()
	c := newCapturingCollector()
	Collect(DEBUG, c)

	log := NewLogger("test")
	log.Debug("message 1")
	if len(c.Captured()) != 1 {
		t.Errorf("Expected first check to find exactly 1 event but found %d instead", len(c.Captured()))
	}
	SetLevel(INFO, c)
	log.Debug("message 2")
	if len(c.Captured()) != 1 {
		t.Errorf("Expected first check to find exactly 1 event but found %d instead", len(c.Captured()))
	}
	SetLevel(DEBUG, c)
	log.Debug("message 3")
	if len(c.Captured()) != 2 {
		t.Errorf("Expected first check to find exactly 2 eventc but found %d instead", len(c.Captured()))
	}
	SetLevel(OFF, c)
	log.Debug("message 4")
	if len(c.Captured()) != 2 {
		t.Errorf("Expected first check to find exactly 2 eventc but found %d instead", len(c.Captured()))
	}
}

func TestSetLevelCollectorNotPresent(t *testing.T) {
	// Make sure nothing blows-up
	defer resetCue()
	c := newCapturingCollector()
	SetLevel(DEBUG, c)
	SetLevel(INFO, c)
}

func TestLoggerString(t *testing.T) {
	defer resetCue()
	log := NewLogger("test")
	s, ok := log.(fmt.Stringer)
	if !ok {
		t.Error("Expected logger type to implement String method")
	}
	_ = s.String()
}

func checkEventExpectation(t *testing.T, event *Event, level Level, message string, err error) {
	if event.Level != level {
		t.Errorf("Invalid event level. Expected: %s, Received: %s", level, event.Level)
	}
	if event.Message != message {
		t.Errorf("Invalid event message. Expected: %s, Received: %s", message, event.Message)
	}
	if event.Error != err {
		t.Errorf("Invalid event error. Expected: %s, Received: %s", err, event.Error)
	}

	ourTestFile := "logger_test.go"
	if !strings.HasSuffix(event.Frames[0].File, ourTestFile) {
		t.Errorf("Invalid frames captured.  Expected source file with suffix %q, but didn't see it", ourTestFile)
	}

	now := time.Now()
	timediff := now.Sub(event.Time)
	if event.Time.IsZero() {
		t.Error("Event has uninitialized time field")
	}
	if timediff == 0 || timediff > 10*time.Minute {
		t.Errorf("Invalid event time.  Expected time diff between now and event time to 0 < diff <= 10 minutes.  Event time: %s, Now: %s, Diff: %s", event.Time.Format(time.Stamp), now.Format(time.Stamp), timediff)
	}
}

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
	"errors"
	"fmt"
	"github.com/bobziuchkovski/cue"
	"time"
)

// We ensure deterministic iteration order by using multiple WithValue
// calls as opposed to a WithFields call.
var ctx = cue.NewContext("test context").
	WithValue("k1", "some value").
	WithValue("k2", 2).
	WithValue("k3", 3.5).
	WithValue("k4", true)

// Test events at all cue event levels.  The *Event instances have 3 frames
// in there Frames field while the *EventNoFrames instances have 0.
var (
	DebugEvent         = GenerateEvent(cue.DEBUG, ctx, "debug event", nil, 3)
	DebugEventNoFrames = GenerateEvent(cue.DEBUG, ctx, "debug event", nil, 0)
	InfoEvent          = GenerateEvent(cue.INFO, ctx, "info event", nil, 3)
	InfoEventNoFrames  = GenerateEvent(cue.INFO, ctx, "info event", nil, 0)
	WarnEvent          = GenerateEvent(cue.WARN, ctx, "warn event", nil, 3)
	WarnEventNoFrames  = GenerateEvent(cue.WARN, ctx, "warn event", nil, 0)
	ErrorEvent         = GenerateEvent(cue.ERROR, ctx, "error event", errors.New("error message"), 3)
	ErrorEventNoFrames = GenerateEvent(cue.ERROR, ctx, "error event", errors.New("error message"), 0)
	FatalEvent         = GenerateEvent(cue.FATAL, ctx, "fatal event", errors.New("fatal message"), 3)
	FatalEventNoFrames = GenerateEvent(cue.FATAL, ctx, "fatal event", errors.New("fatal message"), 0)
)

// GenerateEvent returns a new event for the given parameters.  The frames
// parameter determines how many frames to attach to the Frames field.  The
// generated frames follow a naming pattern based on their index.  The time
// used for the generated event is the same time used by the time package to
// represent time formats.
func GenerateEvent(level cue.Level, context cue.Context, message string, err error, frames int) *cue.Event {
	event := &cue.Event{
		Time:    testTime(),
		Level:   level,
		Context: context,
		Message: message,
		Error:   err,
	}
	for i := frames; i > 0; i-- {
		event.Frames = append(event.Frames, &cue.Frame{
			Package:  fmt.Sprintf("github.com/bobziuchkovski/cue/frame%d", i),
			Function: fmt.Sprintf("github.com/bobziuchkovski/cue/frame%d.function%d", i, i),
			File:     fmt.Sprintf("/path/github.com/bobziuchkovski/cue/frame%d/file%d.go", i, i),
			Line:     i,
		})
	}
	return event
}

func testTime() time.Time {
	t, err := time.Parse(time.RFC822, time.RFC822)
	if err != nil {
		panic(err)
	}
	return t
}

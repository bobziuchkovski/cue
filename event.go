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
	"runtime"
	"time"
)

// Event represents a log event.  A single Event pointer is passed to all
// matching collectors across multiple goroutines.  For this reason, Event
// fields -must not- be altered in place.
type Event struct {
	Time    time.Time // Local time when the event was generated
	Level   Level     // Event severity level
	Context Context   // Context of the logger that generated the event
	Frames  []*Frame  // Stack frames for the call site, or nil if disabled
	Error   error     // The error associated with the message (ERROR and FATAL levels only)
	Message string    // The log message
}

func newEvent(context Context, level Level, cause error, message string) *Event {
	now := time.Now()
	return &Event{
		Time:    now,
		Level:   level,
		Context: context,
		Error:   cause,
		Message: message,
	}
}

func newEventf(context Context, level Level, cause error, format string, values ...interface{}) *Event {
	now := time.Now()
	return &Event{
		Time:    now,
		Level:   level,
		Context: context,
		Error:   cause,
		Message: fmt.Sprintf(format, values...),
	}
}

func (e *Event) captureFrames(skip int, depth int, errorDepth int, recovering bool) {
	skip++
	if e.Level == ERROR || e.Level == FATAL {
		depth = errorDepth
	}
	if depth <= 0 {
		return
	}

	frameFunc := getFrames
	if recovering {
		frameFunc = getRecoveryFrames
	}
	frameptrs := frameFunc(skip, depth)
	if frameptrs == nil {
		return
	}
	e.Frames = make([]*Frame, len(frameptrs))
	for i, ptr := range frameptrs {
		e.Frames[i] = frameForPC(ptr)
	}
}

// Calling panic() adds additional frames to the call stack, so we need to
// find and skip those additional frames.
func getRecoveryFrames(skip int, depth int) []uintptr {
	skip++
	panicFrames := getFrames(skip, maxPanicDepth)
	for i, pc := range panicFrames {
		if frameForPC(pc).Function == "runtime.gopanic" {
			return getFrames(skip+i+1, depth)
		}
	}

	// Couldn't determine the panic frames, so return all the frames, panic
	// included.
	return getFrames(skip, depth)
}

func getFrames(skip int, depth int) []uintptr {
	skip++
	stack := make([]uintptr, depth)
	count := runtime.Callers(skip, stack)
	stack = stack[:count]
	if count > 0 {
		// Per runtime package docs, we need to adjust the pc value in the
		// nearest frame to get the actual caller.
		stack[0]--
	}
	return stack
}

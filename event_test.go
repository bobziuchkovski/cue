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
	"testing"
)

func TestEventSource(t *testing.T) {
	e := &Event{}
	e.captureFrames(1, 1, 1, false)
	if e.Frames[0].Function != "github.com/remerge/cue.TestEventSource" {
		t.Errorf("Event source function doesn't match expectations.  Expected: %s, received: %s", "github.com/remerge/cue.TestEventSource", e.Frames[0].Function)
	}
}

func TestEventStack(t *testing.T) {
	e := &Event{}
	e.captureFrames(1, 2, 2, false)
	if e.Frames[0].Function != "github.com/remerge/cue.TestEventStack" {
		t.Errorf("Event stack[0] function doesn't match expectations.  Expected: %s, received: %s", "github.com/remerge/cue.TestEventStack", e.Frames[0].Function)
	}
	if len(e.Frames) != 2 {
		t.Errorf("Expected 2 frames but received %d instead", len(e.Frames))
	}

	e2 := &Event{}
	if e2.Frames != nil {
		t.Error("Expected Event.Frames to return nil when no frames are captured")
	}
}

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
	"runtime"
	"strings"
	"testing"
)

// This test should be first since it's sensitive to it's own position in the
// source file.
func TestFrameLine(t *testing.T) {
	// The line number we capture/compare is for the next line
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		t.Error("Failed to get current stack pointer")
	}
	frame := frameForPC(pc)
	if frame.Line != 33 {
		t.Errorf("Expected line number 33 but received %d instead", frame.Line)
	}
}

func TestFrameFile(t *testing.T) {
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		t.Error("Failed to get current stack pointer")
	}
	frame := frameForPC(pc)
	if !strings.HasSuffix(frame.File, "github.com/bobziuchkovski/cue/frame_test.go") {
		t.Errorf("Expected frame.File() to have suffix with current file name, but it didn't.  frame.File: %s", frame.File)
	}
}

func TestFrameFunction(t *testing.T) {
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		t.Error("Failed to get current stack pointer")
	}
	frame := frameForPC(pc)
	if frame.Function != "github.com/bobziuchkovski/cue.TestFrameFunction" {
		t.Errorf("Frame function is incorrect.  Expected: %s, Received: %s", "github.com/bobziuchkovski/cue.TestFrameFunction", frame.Function)
	}
}

func TestFramePackage(t *testing.T) {
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		t.Error("Failed to get current stack pointer")
	}
	frame := frameForPC(pc)
	if frame.Package != "github.com/bobziuchkovski/cue" {
		t.Errorf("Frame package is incorrect.  Expected: %s, Received: %s", "github.com/bobziuchkovski/cue", frame.Package)
	}
}

func TestNilFrame(t *testing.T) {
	frame := frameForPC(0)
	if frame.File != UnknownFile {
		t.Errorf("Expected Frame.File to return %q when frame is unknown", UnknownFile)
	}
	if frame.Function != UnknownFunction {
		t.Errorf("Expected Frame.Function to return %q when frame is unknown", UnknownFunction)
	}
	if frame.Line != 0 {
		t.Error("Expected Frame.Line to return 0 when frame is unknown")
	}
	if frame.Package != UnknownPackage {
		t.Errorf("Expected Frame.Package to return %q when frame is unknown", UnknownPackage)
	}
}

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
)

// maxPanicDepth is the maximum number of frames to search when locating the
// call site that triggered panic().  8 frames is arbitrary, but it's
// relatively safe to assume that panic adds less than 8 frames to the stack.
// On amd64 with go 1.6, panic adds 2 frames to the stack.
const maxPanicDepth = 8

var _, _, _, canDetect = runtime.Caller(0)

func doPanic(cause interface{}) {
	panic(cause)
}

// Detect whether the current stack is a panic caused by us.
func ourPanic() bool {
	if !canDetect {
		return false
	}

	framebuf := make([]uintptr, maxPanicDepth)
	copied := runtime.Callers(0, framebuf)
	framebuf = framebuf[:copied]
	for _, pc := range framebuf {
		if frameForPC(pc).Function == "github.com/remerge/cue.doPanic" {
			return true
		}
	}
	return false
}

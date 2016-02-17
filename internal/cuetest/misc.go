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
	"github.com/bobziuchkovski/cue"
	"io"
	"time"
)

// CloseCollector calls c.Close() if c implements the io.Closer interface.
// If c.Close() doesn't return within 5 seconds, CloseCollector panics.
func CloseCollector(c cue.Collector) {
	closer, ok := c.(io.Closer)
	if !ok {
		return
	}

	timer := time.AfterFunc(5*time.Second, func() {
		panic("Failed to close collector within 5 seconds")
	})
	closer.Close()
	timer.Stop()
}

// ResetCue calls cue.Close(time.Minute).  If that returns a non-nil result,
// ResetCue panics.
func ResetCue() {
	err := cue.Close(time.Minute)
	if err != nil {
		panic("Cue failed to reset within a minute")
	}
}

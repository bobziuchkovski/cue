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

func TestOurPanic(t *testing.T) {
	builtinCause := "built-in"
	panicCause := "logger panic"
	panicfCause := "logger panicf"

	regularPanic := func() { panic(builtinCause) }
	logPanic := func() { NewLogger("logger panic").Panic(panicCause, "logger panic") }
	logPanicf := func() { NewLogger("logger panicf").Panicf(panicfCause, "logger %s", "panicf") }
	if recoverAndCheckOurPanic(regularPanic) {
		t.Error("Regular built-in panic incorrectly detected as our own.")
	}
	if !recoverAndCheckOurPanic(logPanic) {
		t.Error("Logger panic not detected as our own")
	}
	if !recoverAndCheckOurPanic(logPanicf) {
		t.Error("Logger panicf not detected as our own")
	}
}

func recoverAndCheckOurPanic(fn func()) (ours bool) {
	defer func() {
		recover()
		ours = ourPanic()
	}()
	fn()
	return
}

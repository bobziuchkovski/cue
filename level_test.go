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

func TestLevelString(t *testing.T) {
	if OFF.String() != "OFF" {
		t.Errorf("OFF.String value is incorrect.  Expected %q but received %q instead", "OFF", OFF.String())
	}
	if DEBUG.String() != "DEBUG" {
		t.Errorf("DEBUG.String value is incorrect.  Expected %q but received %q instead", "DEBUG", DEBUG.String())
	}
	if INFO.String() != "INFO" {
		t.Errorf("INFO.String value is incorrect.  Expected %q but received %q instead", "INFO", INFO.String())
	}
	if WARN.String() != "WARN" {
		t.Errorf("WARN.String value is incorrect.  Expected %q but received %q instead", "WARN", WARN.String())
	}
	if ERROR.String() != "ERROR" {
		t.Errorf("ERROR.String value is incorrect.  Expected %q but received %q instead", "ERROR", ERROR.String())
	}
	if FATAL.String() != "FATAL" {
		t.Errorf("FATAL.String value is incorrect.  Expected %q but received %q instead", "FATAL", FATAL.String())
	}
	if Level(42).String() != "INVALID LEVEL" {
		t.Error("Expected to see INVALID LEVEL for bogus level")
	}
}

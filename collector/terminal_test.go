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

package collector

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bobziuchkovski/cue/internal/cuetest"
)

const terminalDebugStr = "Jan  2 15:04:00 DEBUG file3.go:3 debug event k1=\"some value\" k2=2 k3=3.5 k4=true\n"
const terminalErrorStr = "Jan  2 15:04:00 ERROR file3.go:3 error event: error message k1=\"some value\" k2=2 k3=3.5 k4=true\n"

func TestTerminal(t *testing.T) {
	realStdout, realStderr := os.Stdout, os.Stderr
	defer restoreStdoutStderr(realStdout, realStderr)

	stdout, _ := replaceStdoutStderr()
	c := Terminal{}.New()

	c.Collect(cuetest.DebugEvent)
	c.Collect(cuetest.ErrorEvent)
	restoreStdoutStderr(realStdout, realStderr)

	err := stdout.Close()
	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}
	checkFileContents(t, stdout.Name(), terminalDebugStr+terminalErrorStr)
}

func TestTerminalStderr(t *testing.T) {
	realStdout, realStderr := os.Stdout, os.Stderr
	defer restoreStdoutStderr(realStdout, realStderr)

	stdout, stderr := replaceStdoutStderr()
	c := Terminal{ErrorsToStderr: true}.New()

	c.Collect(cuetest.DebugEvent)
	c.Collect(cuetest.ErrorEvent)
	restoreStdoutStderr(realStdout, realStderr)

	err := stdout.Close()
	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}
	err = stderr.Close()
	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}
	checkFileContents(t, stdout.Name(), terminalDebugStr)
	checkFileContents(t, stderr.Name(), terminalErrorStr)
}

func TestTerminalString(t *testing.T) {
	c := Terminal{ErrorsToStderr: true}.New()

	// Ensure nothing panics
	_ = fmt.Sprint(c)
}

func restoreStdoutStderr(stdout, stderr *os.File) {
	os.Stdout = stdout
	os.Stderr = stderr
}

func replaceStdoutStderr() (stdout, stderr *os.File) {
	var err error
	stdout, err = ioutil.TempFile("", "test-cue")
	if err != nil {
		panic(err)
	}
	os.Stdout = stdout

	stderr, err = ioutil.TempFile("", "test-cue")
	if err != nil {
		panic(err)
	}
	os.Stderr = stderr

	return
}

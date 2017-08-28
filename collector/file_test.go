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
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/remerge/cue/format"
	"github.com/remerge/cue/internal/cuetest"
)

const fileEventStr = "Jan  2 15:04:00 DEBUG file3.go:3 debug event k1=\"some value\" k2=2 k3=3.5 k4=true\n"

func TestFileNilCollector(t *testing.T) {
	c := File{}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the file path is missing, but got %s instead", c)
	}
}

func TestFile(t *testing.T) {
	tmp := tmpDir()
	defer os.RemoveAll(tmp)

	file := path.Join(tmp, "file")
	c := File{Path: file}.New()
	c.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c)
	checkFileContents(t, file, fileEventStr)
}

func TestFileDefaultOptions(t *testing.T) {
	tmp := tmpDir()
	defer os.RemoveAll(tmp)

	file := path.Join(tmp, "file")
	opts := File{Path: file}

	c1 := opts.New()
	c1.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c1)
	checkFileContents(t, file, fileEventStr)

	// Check appending
	c2 := opts.New()
	c2.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c2)
	checkFileContents(t, file, fileEventStr+fileEventStr)

	// Check file mode
	stat, err := os.Stat(file)
	if err != nil {
		t.Errorf("Encountered unexpected error stat'ing file: %s", err)
	}
	if stat.Mode() != 0600 {
		t.Errorf("Expected file mode of %s, but got %s instead", os.FileMode(0600), stat.Mode())
	}
}

func TestFileExplicitOptions(t *testing.T) {
	tmp := tmpDir()
	defer os.RemoveAll(tmp)

	file := path.Join(tmp, "file")
	opts := File{
		Path:      file,
		Flags:     os.O_CREATE | os.O_WRONLY,
		Perms:     0640,
		Formatter: format.HumanMessage,
	}

	// Ensure custom formatter is used
	c1 := opts.New()
	c1.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c1)
	checkFileContents(t, file, "debug event k1=\"some value\" k2=2 k3=3.5 k4=true\n")

	// Ensure file is recreated (no append flag specified)
	c2 := opts.New()
	c2.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c2)
	checkFileContents(t, file, "debug event k1=\"some value\" k2=2 k3=3.5 k4=true\n")

	// Check file mode
	stat, err := os.Stat(file)
	if err != nil {
		t.Errorf("Encountered unexpected error stat'ing file: %s", err)
	}
	if stat.Mode() != 0640 {
		t.Errorf("Expected file mode of %s, but got %s instead", os.FileMode(0640), stat.Mode())
	}
}

func TestFileReopenOnError(t *testing.T) {
	tmp := tmpDir()
	defer os.RemoveAll(tmp)

	file := path.Join(tmp, "nonexistant", "file")
	opts := File{Path: file}

	c1 := opts.New()
	err := c1.Collect(cuetest.DebugEvent)
	if err == nil {
		t.Error("Expected to receive error when directory doesn't exist for file, but didn't")
	}

	err = os.MkdirAll(path.Join(tmp, "nonexistant"), 0700)
	if err != nil {
		t.Errorf("Encountered unexpected error creating directory: %s", err)
	}

	err = c1.Collect(cuetest.DebugEvent)
	if err != nil {
		t.Errorf("Encountered unexpected error writing to file, even though directory now exists: %s", err)
	}

	cuetest.CloseCollector(c1)
	checkFileContents(t, file, fileEventStr)
}

func TestFileReopenSignal(t *testing.T) {
	tmp := tmpDir()
	defer os.RemoveAll(tmp)

	file := path.Join(tmp, "file")
	c := File{
		Path:         file,
		ReopenSignal: syscall.SIGHUP,
	}.New()
	c.Collect(cuetest.DebugEvent)

	// Remove the opened log file
	err := os.Remove(file)
	if err != nil {
		t.Errorf("Encountered unexpected error removing file: %s", err)
	}

	// Send SIGHUP to ourselves
	pid := os.Getpid()
	proc, err := os.FindProcess(pid)
	if err != nil {
		t.Error("Failed to get our pid")
	}
	proc.Signal(syscall.SIGHUP)

	// Wait for reopen to occurm which will recreate the file
	waitExists(file, 5*time.Second)

	c.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c)
	checkFileContents(t, file, fileEventStr)
}

func TestFileReopenMissing(t *testing.T) {
	tmp := tmpDir()
	defer os.RemoveAll(tmp)

	file := path.Join(tmp, "file")
	c := File{
		Path:          file,
		ReopenMissing: time.Millisecond,
	}.New()
	c.Collect(cuetest.DebugEvent)

	err := os.Remove(file)
	if err != nil {
		t.Errorf("Encountered unexpected error removing file: %s", err)
	}
	waitExists(file, 5*time.Second)

	c.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c)
	checkFileContents(t, file, fileEventStr)
}

func TestFileString(t *testing.T) {
	tmp := tmpDir()
	defer os.RemoveAll(tmp)

	file := path.Join(tmp, "file")
	c := File{Path: file}.New()

	// Ensure nothing panics
	_ = fmt.Sprint(c)
}

func tmpDir() string {
	dir, err := ioutil.TempDir("", "cue-test")
	if err != nil {
		panic(err)
	}
	return dir
}

func waitExists(path string, timeout time.Duration) {
	timer := time.AfterFunc(timeout, func() {
		panic("timeout waiting for file to exist")
	})
	for {
		_, err := os.Stat(path)
		if err == nil {
			timer.Stop()
			return
		}
	}
}

func checkFileContents(t *testing.T, path string, expected string) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Errorf("Encountered unexpected error reading file contents: %s", err)
	}

	if string(bytes) != expected {
		t.Errorf(`File contents don't match expectations

Expected
========
%q

Received
========
%q`, expected, string(bytes))
	}
}

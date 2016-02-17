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
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/format"
	"os"
	"os/signal"
	"sync"
	"time"
)

// File represents configuration for file-based Collector instances. The default
// settings create/append to a file at the given path. File rotation is not
// and will not be supported, but the ReopenSignal and ReopenMissing params
// may be used to coordinate with external log rotators.
type File struct {
	// Required
	Path string

	// Optional
	Flags     int              // Default: os.O_CREATE | os.O_WRONLY | os.O_APPEND
	Perms     os.FileMode      // Default: 0600
	Formatter format.Formatter // Default: format.HumanReadable

	// If set, reopen the file if the specified signal is received.  On Unix
	// SIGHUP is often used for this purpose.
	ReopenSignal os.Signal

	// If set, reopen the file if it's missing.  The file path will be checked
	// at the time interval specified.
	ReopenMissing time.Duration
}

// New returns a new collector based on the File configuration.
func (f File) New() cue.Collector {
	if f.Path == "" {
		log.Warn("File.New called to created a collector, but Path param is empty.  Returning nil collector.")
		return nil
	}
	if f.Formatter == nil {
		f.Formatter = format.HumanReadable
	}
	if f.Flags == 0 {
		f.Flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}
	if f.Perms == 0 {
		f.Perms = 0600
	}

	fc := &fileCollector{File: f}
	fc.watchSignal()
	fc.watchRemoval()
	return fc
}

type fileCollector struct {
	File

	mu     sync.Mutex
	file   *os.File
	opened bool
}

func (f *fileCollector) String() string {
	return fmt.Sprintf("File(path=%s)", f.Path)
}

func (f *fileCollector) Collect(event *cue.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	err := f.ensureOpen()
	if err != nil {
		f.ensureClosed()
		return err
	}

	buf := format.GetBuffer()
	defer format.ReleaseBuffer(buf)
	f.Formatter(buf, event)

	bytes := buf.Bytes()
	if bytes[len(bytes)-1] != byte('\n') {
		bytes = append(bytes, byte('\n'))
	}
	_, err = f.file.Write(bytes)
	if err != nil {
		f.ensureClosed()
	}
	return err
}

func (f *fileCollector) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.file != nil {
		return f.file.Close()
	}
	return nil
}

func (f *fileCollector) reopen() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ensureClosed()
	return f.ensureOpen()
}

func (f *fileCollector) ensureOpen() error {
	if f.file != nil {
		return nil
	}

	var err error
	f.file, err = os.OpenFile(f.Path, f.Flags, f.Perms)
	if err == nil {
		f.opened = true
	}
	return err
}

func (f *fileCollector) ensureClosed() {
	if f != nil {
		f.file.Close()
		f.file = nil
	}
	f.opened = false
}

func (f *fileCollector) watchSignal() {
	if f.ReopenSignal == nil {
		return
	}
	triggered := make(chan os.Signal, 1)
	signal.Notify(triggered, f.ReopenSignal)

	go func() {
		for {
			<-triggered
			f.reopen()
		}
	}()
}

func (f *fileCollector) watchRemoval() {
	if f.ReopenMissing == 0 {
		return
	}
	go func() {
		for {
			time.Sleep(f.ReopenMissing)
			_, err := os.Stat(f.Path)
			if os.IsNotExist(err) {
				f.reopen()
			}
		}
	}()
}

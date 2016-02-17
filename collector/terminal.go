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
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/format"
	"os"
)

// Terminal represents configuration for stdout/stderr collection.  By
// default, all events are logged to stdout.
type Terminal struct {
	Formatter      format.Formatter // Default: format.HumanReadable
	ErrorsToStderr bool             // If set, ERROR and FATAL events are written to stderr
}

// New returns a new collector based on the Terminal configuration.
func (t Terminal) New() cue.Collector {
	if t.Formatter == nil {
		t.Formatter = format.HumanReadable
	}
	return &terminalCollector{Terminal: t}
}

type terminalCollector struct {
	Terminal
}

func (t *terminalCollector) String() string {
	return "Terminal()"
}

func (t *terminalCollector) Collect(event *cue.Event) error {
	output := os.Stdout
	if t.ErrorsToStderr && (event.Level == cue.ERROR || event.Level == cue.FATAL) {
		output = os.Stderr
	}

	buf := format.GetBuffer()
	defer format.ReleaseBuffer(buf)
	t.Formatter(buf, event)

	bytes := buf.Bytes()
	if bytes[len(bytes)-1] != byte('\n') {
		bytes = append(bytes, byte('\n'))
	}

	_, err := output.Write(bytes)
	return err
}

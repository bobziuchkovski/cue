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
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/bobziuchkovski/cue/internal/cuetest"
)

const socketEventStr = "Jan  2 15:04:00 DEBUG file3.go:3 debug event k1=\"some value\" k2=2 k3=3.5 k4=true"

func TestSocketNilCollector(t *testing.T) {
	c := Socket{Address: "bogus"}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the socket network is missing, but got %s instead", c)
	}

	c = Socket{Network: "bogus"}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the socket address is missing, but got %s instead", c)
	}
}

func TestSocketBasic(t *testing.T) {
	recorder := cuetest.NewTCPRecorder()
	recorder.Start()
	defer recorder.Close()

	c := Socket{
		Network: "tcp",
		Address: recorder.Address(),
	}.New()

	c.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c)
	recorder.CheckStringContents(t, socketEventStr)
}

func TestSocketTLS(t *testing.T) {
	recorder := cuetest.NewTLSRecorder()
	recorder.Start()
	defer recorder.Close()

	c := Socket{
		Network: "tcp",
		Address: recorder.Address(),
		TLS:     &tls.Config{InsecureSkipVerify: true},
	}.New()

	c.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c)
	recorder.CheckStringContents(t, socketEventStr)
}

func TestSocketReopenOnError(t *testing.T) {
	recorder := cuetest.NewTCPRecorder()
	defer recorder.Close()

	c := Socket{
		Network: "tcp",
		Address: recorder.Address(),
	}.New()

	err := c.Collect(cuetest.DebugEvent)
	if err == nil {
		t.Error("Expected to see a collector error but didn't")
	}

	recorder.Start()
	err = c.Collect(cuetest.DebugEvent)
	if err != nil {
		t.Errorf("Encountered unexpected collector error: %s", err)
	}

	cuetest.CloseCollector(c)
	recorder.CheckStringContents(t, socketEventStr)
}

func TestSocketString(t *testing.T) {
	recorder := cuetest.NewTCPRecorder()
	defer recorder.Close()

	c := Socket{
		Network: "tcp",
		Address: recorder.Address(),
	}.New()

	// Ensure nothing panics
	_ = fmt.Sprint(c)
}

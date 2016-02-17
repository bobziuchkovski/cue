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
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/internal/cuetest"
	"os"
	"regexp"
	"testing"
)

func TestSyslogNilCollector(t *testing.T) {
	c := Syslog{}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the app is missing, but got %s instead", c)
	}
}

func TestSyslogLocalCollector(t *testing.T) {
	var localPath string
	for _, path := range syslogSockets {
		_, err := os.Stat(path)
		if err == nil {
			localPath = path
		}
	}

	c := Syslog{App: "testapp"}.New()
	switch {
	case localPath != "" && c == nil:
		t.Errorf("Found local syslog socket %s but Syslog{App: %q}.New() returned nil collector. ", localPath, "testapp")
	case localPath == "" && c != nil:
		t.Errorf("No local syslog socket found.  Expected Syslog{App: %q}.New() to return nil collector but it returned %s instead", "testapp", c)
	}
}

func TestSyslogDebug(t *testing.T) {
	testSyslogEvent(t, cuetest.DebugEvent)
}

func TestSyslogInfo(t *testing.T) {
	testSyslogEvent(t, cuetest.InfoEvent)
}

func TestSyslogWarn(t *testing.T) {
	testSyslogEvent(t, cuetest.WarnEvent)
}

func TestSyslogError(t *testing.T) {
	testSyslogEvent(t, cuetest.ErrorEvent)
}

func TestSyslogFatal(t *testing.T) {
	testSyslogEvent(t, cuetest.FatalEvent)
}

func testSyslogEvent(t *testing.T, event *cue.Event) {
	recorder := cuetest.NewTCPRecorder()
	recorder.Start()
	defer recorder.Close()

	c := Syslog{
		App:      "testapp",
		Facility: LOCAL4,
		Network:  "tcp",
		Address:  recorder.Address(),
	}.New()

	c.Collect(event)
	cuetest.CloseCollector(c)
	checkSyslogContents(t, "testapp", LOCAL4, string(recorder.Contents()), event)
}

func TestSyslogTLS(t *testing.T) {
	recorder := cuetest.NewTLSRecorder()
	recorder.Start()
	defer recorder.Close()

	c := Syslog{
		App:      "testapp",
		Facility: LOCAL4,
		Network:  "tcp",
		Address:  recorder.Address(),
		TLS:      &tls.Config{InsecureSkipVerify: true},
	}.New()

	c.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c)
	checkSyslogContents(t, "testapp", LOCAL4, string(recorder.Contents()), cuetest.DebugEvent)
}

func TestSyslogString(t *testing.T) {
	recorder := cuetest.NewTCPRecorder()
	recorder.Start()
	defer recorder.Close()

	c := Syslog{
		App:      "testapp",
		Facility: LOCAL4,
		Network:  "tcp",
		Address:  recorder.Address(),
	}.New()

	// Ensure nothing panics
	_ = fmt.Sprint(c)
}

func TestStructuredSyslogNilCollector(t *testing.T) {
	c := StructuredSyslog{}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the app is missing, but got %s instead", c)
	}
}

func TestStructuredSyslogLocalCollector(t *testing.T) {
	var localPath string
	for _, path := range syslogSockets {
		_, err := os.Stat(path)
		if err == nil {
			localPath = path
		}
	}

	c := StructuredSyslog{App: "testapp"}.New()
	switch {
	case localPath != "" && c == nil:
		t.Errorf("Found local syslog socket %s but StructuredSyslog{App: %q}.New() returned nil collector. ", localPath, "testapp")
	case localPath == "" && c != nil:
		t.Errorf("No local syslog socket found.  Expected StructuredSyslog{App: %q}.New() to return nil collector but it returned %s instead", "testapp", c)
	}
}

func TestStructuredSyslogDebug(t *testing.T) {
	testStructuredSyslogEvent(t, cuetest.DebugEvent)
}

func TestStructuredSyslogInfo(t *testing.T) {
	testStructuredSyslogEvent(t, cuetest.InfoEvent)
}

func TestStructuredSyslogWarn(t *testing.T) {
	testStructuredSyslogEvent(t, cuetest.WarnEvent)
}

func TestStructuredSyslogError(t *testing.T) {
	testStructuredSyslogEvent(t, cuetest.ErrorEvent)
}

func TestStructuredSyslogFatal(t *testing.T) {
	testStructuredSyslogEvent(t, cuetest.FatalEvent)
}

func testStructuredSyslogEvent(t *testing.T, event *cue.Event) {
	recorder := cuetest.NewTCPRecorder()
	recorder.Start()
	defer recorder.Close()

	c := StructuredSyslog{
		App:      "testapp",
		Facility: LOCAL4,
		Network:  "tcp",
		Address:  recorder.Address(),
		ID:       "test@12345",
	}.New()

	c.Collect(event)
	cuetest.CloseCollector(c)
	checkStructuredSyslogContents(t, "testapp", LOCAL4, "test@12345", string(recorder.Contents()), event)
}

func TestStructuredSyslogTLS(t *testing.T) {
	recorder := cuetest.NewTLSRecorder()
	recorder.Start()
	defer recorder.Close()

	c := StructuredSyslog{
		App:      "testapp",
		Facility: LOCAL4,
		Network:  "tcp",
		Address:  recorder.Address(),
		ID:       "test@12345",
		TLS:      &tls.Config{InsecureSkipVerify: true},
	}.New()

	c.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c)
	checkStructuredSyslogContents(t, "testapp", LOCAL4, "test@12345", string(recorder.Contents()), cuetest.DebugEvent)
}

func TestStructuredSyslogByteOrderMark(t *testing.T) {
	recorder := cuetest.NewTCPRecorder()
	recorder.Start()
	defer recorder.Close()

	c := StructuredSyslog{
		App:      "testapp",
		Facility: LOCAL4,
		Network:  "tcp",
		Address:  recorder.Address(),
		ID:       "test@12345",
		WriteBOM: true,
	}.New()

	c.Collect(cuetest.DebugEvent)
	cuetest.CloseCollector(c)
	if !bytes.Contains(recorder.Contents(), []byte{0xef, 0xbb, 0xbf}) {
		t.Error("Expected to see byte order mark (BOM) in the output but didn't")
	}
}

func TestStructuredSyslogString(t *testing.T) {
	recorder := cuetest.NewTCPRecorder()
	recorder.Start()
	defer recorder.Close()

	c := StructuredSyslog{
		App:      "testapp",
		Facility: LOCAL4,
		Network:  "tcp",
		Address:  recorder.Address(),
	}.New()

	// Ensure nothing panics
	_ = fmt.Sprint(c)
}

func TestFacilityString(t *testing.T) {
	facilities := []Facility{
		KERN, USER, MAIL, DAEMON, AUTH, SYSLOG, LPR, NEWS, UUCP, CRON, AUTHPRIV, FTP, NTP, AUDIT, ALERT,
		LOCAL0, LOCAL1, LOCAL2, LOCAL3, LOCAL4, LOCAL5, LOCAL6, LOCAL7,
	}

	// Ensure nothing panics
	for _, f := range facilities {
		_ = fmt.Sprint(f)
	}
	_ = fmt.Sprint(Facility(1000))
}

func checkSyslogContents(t *testing.T, app string, facility Facility, content string, event *cue.Event) {
	pri := 8*int(facility) + int(severityFor(event.Level))
	pattern := fmt.Sprintf("^<%d>2006-01-02T15:04:00-\\d{2}:\\d{2} \\S+ %s\\[\\d+\\]:[^\\n]*\\n$", pri, app)
	re := regexp.MustCompile(pattern)
	if !re.MatchString(content) {
		t.Errorf("Content %q doesn't match pattern %q", content, pattern)
	}
}

func checkStructuredSyslogContents(t *testing.T, app string, facility Facility, id string, content string, event *cue.Event) {
	pri := 8*int(facility) + int(severityFor(event.Level))
	pattern := fmt.Sprintf("^<%d>1 2006-01-02T15:04:00.000000-\\d{2}:\\d{2} \\S+ %s %s\\[\\d+\\] - \\[%s[^\\n]*?\\][^\\n]*\\n$", pri, app, app, id)
	re := regexp.MustCompile(pattern)
	if !re.MatchString(content) {
		t.Errorf("Content %q doesn't match pattern %q", content, pattern)
	}
}

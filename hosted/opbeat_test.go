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

package hosted

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/remerge/cue"
	"github.com/remerge/cue/internal/cuetest"
)

const opbeatJSON = `
{
  "culprit": "github.com/remerge/cue/frame3.function3",
  "exception": {
    "module": "github.com/remerge/cue/frame3",
    "type": "errors.errorString",
    "value": "error message"
  },
  "extra": {
    "extra": "extra value",
    "k1": "some value",
    "k2": 2,
    "k3": 3.5,
    "k4": true
  },
  "level": "error",
  "logger": "test context",
  "machine": {
    "hostname": "pegasus.bobbyz.org"
  },
  "message": "error event: error message",
  "stacktrace": {
    "frames": [
      {
        "filename": "/path/github.com/remerge/cue/frame1/file1.go",
        "function": "github.com/remerge/cue/frame1.function1",
        "lineno": 1
      },
      {
        "filename": "/path/github.com/remerge/cue/frame2/file2.go",
        "function": "github.com/remerge/cue/frame2.function2",
        "lineno": 2
      },
      {
        "filename": "/path/github.com/remerge/cue/frame3/file3.go",
        "function": "github.com/remerge/cue/frame3.function3",
        "lineno": 3
      }
    ]
  },
  "timestamp": "2006-01-02T22:04:00.000Z"
}
`

const opbeatNoFramesJSON = `
{
  "exception": {
    "type": "errors.errorString",
    "value": "error message"
  },
  "extra": {
    "extra": "extra value",
    "k1": "some value",
    "k2": 2,
    "k3": 3.5,
    "k4": true
  },
  "level": "error",
  "logger": "test context",
  "machine": {
    "hostname": "pegasus.bobbyz.org"
  },
  "message": "error event: error message",
  "timestamp": "2006-01-02T22:04:00.000Z"
}
`

func TestOpbeatNilCollector(t *testing.T) {
	c := Opbeat{}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the token is missing, but got %s instead", c)
	}
}

func TestOpbeat(t *testing.T) {
	checkOpbeatEvent(t, cuetest.ErrorEvent, opbeatJSON)
}

func TestOpbeatNoFrames(t *testing.T) {
	checkOpbeatEvent(t, cuetest.ErrorEventNoFrames, opbeatNoFramesJSON)
}

func TestOpbeatString(t *testing.T) {
	_ = fmt.Sprint(getOpbeatCollector())
}

func TestOpbeatLevels(t *testing.T) {
	m := map[cue.Level]string{
		cue.DEBUG: "debug",
		cue.INFO:  "info",
		cue.WARN:  "warning",
		cue.ERROR: "error",
		cue.FATAL: "fatal",
	}
	for k, v := range m {
		if opbeatLevel(k) != v {
			t.Errorf("Expected cue level %q to map to opbeat level %q but it didn't", k, v)
		}
	}
}

func checkOpbeatEvent(t *testing.T, event *cue.Event, expected string) {
	req, err := getOpbeatCollector().formatRequest(event)
	if err != nil {
		t.Errorf("Encountered unexpected error formatting http request: %s", err)
	}
	requestJSON := cuetest.ParseRequestJSON(req)
	expectedJSON := cuetest.ParseStringJSON(expected)

	if cuetest.NestedFetch(requestJSON, "machine", "hostname") == "!(MISSING)" {
		t.Error("Hostname is missing from request")
	}
	if cuetest.NestedFetch(requestJSON, "timestamp") == "!(MISSING)" {
		t.Error("Timestamp is missing from request")
	}

	cuetest.NestedDelete(requestJSON, "machine", "hostname")
	cuetest.NestedDelete(expectedJSON, "machine", "hostname")
	cuetest.NestedDelete(requestJSON, "timestamp")
	cuetest.NestedDelete(expectedJSON, "timestamp")
	cuetest.NestedCompare(t, requestJSON, expectedJSON)
}

func getOpbeatCollector() *opbeatCollector {
	c := Opbeat{
		Token:          "test",
		AppID:          "app",
		OrganizationID: "org",
		ExtraContext:   cue.NewContext("extra").WithValue("extra", "extra value"),
	}.New()
	oc, ok := c.(*opbeatCollector)
	if !ok {
		panic(fmt.Sprintf("Expected to see a *opbeatCollector but got %s instead", reflect.TypeOf(c)))
	}
	return oc
}

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

const rollbarJSON = `
{
  "access_token": "test",
  "data": {
    "body": {
      "trace": {
        "exception": {
          "class": "errors.errorString",
          "description": "error event: error message",
          "message": "error message"
        },
        "frames": [
          {
            "filename": "/path/github.com/remerge/cue/frame1/file1.go",
            "lineno": 1,
            "method": "github.com/remerge/cue/frame1.function1"
          },
          {
            "filename": "/path/github.com/remerge/cue/frame2/file2.go",
            "lineno": 2,
            "method": "github.com/remerge/cue/frame2.function2"
          },
          {
            "filename": "/path/github.com/remerge/cue/frame3/file3.go",
            "lineno": 3,
            "method": "github.com/remerge/cue/frame3.function3"
          }
        ]
      }
    },
    "code_version": "1.2.3",
    "custom": {
      "extra": "extra value",
      "k1": "some value",
      "k2": 2,
      "k3": 3.5,
      "k4": true
    },
    "environment": "test",
    "framework": "sliced-bread",
    "language": "go",
    "level": "error",
    "notifier": {
      "name": "github.com/remerge/cue",
      "version": "0.7.0"
    },
    "platform": "darwin",
    "server": {
      "host": "pegasus.bobbyz.org"
    },
    "timestamp": 1136239440
  }
}
`

const rollbarNoFramesJSON = `
{
  "access_token": "test",
  "data": {
    "body": {
      "message": {
        "body": "error event: error message"
      }
    },
    "code_version": "1.2.3",
    "custom": {
      "extra": "extra value",
      "k1": "some value",
      "k2": 2,
      "k3": 3.5,
      "k4": true
    },
    "environment": "test",
    "framework": "sliced-bread",
    "language": "go",
    "level": "error",
    "notifier": {
      "name": "github.com/remerge/cue",
      "version": "0.7.0"
    },
    "platform": "darwin",
    "server": {
      "host": "pegasus.bobbyz.org"
    },
    "timestamp": 1136239440
  }
}
`

func TestRollbarNilCollector(t *testing.T) {
	c := Rollbar{}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the token is missing, but got %s instead", c)
	}
}

func TestRollbar(t *testing.T) {
	checkRollbarEvent(t, cuetest.ErrorEvent, rollbarJSON)
}

func TestRollbarNoFrames(t *testing.T) {
	checkRollbarEvent(t, cuetest.ErrorEventNoFrames, rollbarNoFramesJSON)
}

func TestRollbarString(t *testing.T) {
	_ = fmt.Sprint(getRollbarCollector())
}

func TestRollbarLevels(t *testing.T) {
	m := map[cue.Level]string{
		cue.DEBUG: "debug",
		cue.INFO:  "info",
		cue.WARN:  "warning",
		cue.ERROR: "error",
		cue.FATAL: "critical",
	}
	for k, v := range m {
		if rollbarLevel(k) != v {
			t.Errorf("Expected cue level %q to map to rollbar level %q but it didn't", k, v)
		}
	}
}

func checkRollbarEvent(t *testing.T, event *cue.Event, expected string) {
	req, err := getRollbarCollector().formatRequest(event)
	if err != nil {
		t.Errorf("Encountered unexpected error formatting http request: %s", err)
	}
	requestJSON := cuetest.ParseRequestJSON(req)
	expectedJSON := cuetest.ParseStringJSON(expected)

	version := cuetest.NestedFetch(requestJSON, "data", "notifier", "version")
	if version != fmt.Sprintf("%d.%d.%d", cue.Version.Major, cue.Version.Minor, cue.Version.Patch) {
		t.Errorf("Invalid notifier version: %s", version)
	}
	if cuetest.NestedFetch(requestJSON, "data", "server", "host") == "!(MISSING)" {
		t.Error("Server host is missing from request")
	}
	if cuetest.NestedFetch(requestJSON, "data", "platform") == "!(MISSING)" {
		t.Error("Platform is missing from request")
	}
	if cuetest.NestedFetch(requestJSON, "data", "timestamp") == "!(MISSING)" {
		t.Error("Timestamp is missing from request")
	}

	cuetest.NestedDelete(requestJSON, "data", "notifier", "version")
	cuetest.NestedDelete(expectedJSON, "data", "notifier", "version")
	cuetest.NestedDelete(requestJSON, "data", "server", "host")
	cuetest.NestedDelete(expectedJSON, "data", "server", "host")
	cuetest.NestedDelete(requestJSON, "data", "platform")
	cuetest.NestedDelete(expectedJSON, "data", "platform")
	cuetest.NestedDelete(requestJSON, "data", "timestamp")
	cuetest.NestedDelete(expectedJSON, "data", "timestamp")
	cuetest.NestedCompare(t, requestJSON, expectedJSON)
}

func getRollbarCollector() *rollbarCollector {
	c := Rollbar{
		Token:            "test",
		Environment:      "test",
		ProjectVersion:   "1.2.3",
		ProjectFramework: "sliced-bread",
		ExtraContext:     cue.NewContext("extra").WithValue("extra", "extra value"),
	}.New()
	rc, ok := c.(*rollbarCollector)
	if !ok {
		panic(fmt.Sprintf("Expected to see a *rollbarCollector but got %s instead", reflect.TypeOf(c)))
	}
	return rc
}

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
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/internal/cuetest"
	"reflect"
	"testing"
)

const sentryJSON = `
{
  "culprit": "github.com/bobziuchkovski/cue/frame3.function3",
  "event_id": "c51c8551adaa42ee9b9fd0b63d462e8d",
  "exception": {
    "module": "github.com/bobziuchkovski/cue/frame3",
    "stacktrace": {
      "frames": [
        {
          "filename": "/path/github.com/bobziuchkovski/cue/frame1/file1.go",
          "function": "github.com/bobziuchkovski/cue/frame1.function1",
          "lineno": 1,
          "module": "github.com/bobziuchkovski/cue/frame1"
        },
        {
          "filename": "/path/github.com/bobziuchkovski/cue/frame2/file2.go",
          "function": "github.com/bobziuchkovski/cue/frame2.function2",
          "lineno": 2,
          "module": "github.com/bobziuchkovski/cue/frame2"
        },
        {
          "filename": "/path/github.com/bobziuchkovski/cue/frame3/file3.go",
          "function": "github.com/bobziuchkovski/cue/frame3.function3",
          "lineno": 3,
          "module": "github.com/bobziuchkovski/cue/frame3"
        }
      ]
    },
    "type": "errors.errorString",
    "value": "error event"
  },
  "level": "error",
  "logger": "test context",
  "message": "error event: error message",
  "platform": "go",
  "server_name": "pegasus.bobbyz.org",
  "tags": [
    [
      "extra",
      "extra value"
    ],
    [
      "k1",
      "some value"
    ],
    [
      "k2",
      "2"
    ],
    [
      "k3",
      "3.5"
    ],
    [
      "k4",
      "true"
    ]
  ],
  "timestamp": "2006-01-02T22:04:00"
}
`

const sentryNoFramesJSON = `
{
  "event_id": "679989e954034f948cc5e5cb220d32aa",
  "exception": {
    "type": "errors.errorString",
    "value": "error event"
  },
  "level": "error",
  "logger": "test context",
  "message": "error event: error message",
  "platform": "go",
  "server_name": "pegasus.bobbyz.org",
  "tags": [
    [
      "extra",
      "extra value"
    ],
    [
      "k1",
      "some value"
    ],
    [
      "k2",
      "2"
    ],
    [
      "k3",
      "3.5"
    ],
    [
      "k4",
      "true"
    ]
  ],
  "timestamp": "2006-01-02T22:04:00"
}
`

func TestSentryNilCollector(t *testing.T) {
	c := Sentry{}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the DSN is missing, but got %s instead", c)
	}
}

func TestSentry(t *testing.T) {
	checkSentryEvent(t, cuetest.ErrorEvent, sentryJSON)
}

func TestSentryNoFrames(t *testing.T) {
	checkSentryEvent(t, cuetest.ErrorEventNoFrames, sentryNoFramesJSON)
}

func TestSentryString(t *testing.T) {
	_ = fmt.Sprint(getSentryCollector())
}

func TestSentryValidDSN(t *testing.T) {
	if validDSN(":bogus") {
		t.Errorf("%q should register as an invalid DSN, but that's not the case", ":bogus")
	}
	if validDSN("http://sentry.private") {
		t.Errorf("%q should register as an invalid DSN due to missing user, but that's not the case", "http://bob@sentry.private")
	}
	if validDSN("http://:pass@sentry.private") {
		t.Errorf("%q should register as an invalid DSN due to missing user, but that's not the case", "http://bob@sentry.private")
	}
	if validDSN("http://bob@sentry.private") {
		t.Errorf("%q should register as an invalid DSN due to missing password, but that's not the case", "http://bob@sentry.private")
	}
}

func TestSentryLevels(t *testing.T) {
	m := map[cue.Level]string{
		cue.DEBUG: "debug",
		cue.INFO:  "info",
		cue.WARN:  "warning",
		cue.ERROR: "error",
		cue.FATAL: "fatal",
	}
	for k, v := range m {
		if sentryLevel(k) != v {
			t.Errorf("Expected cue level %q to map to sentry level %q but it didn't", k, v)
		}
	}
}

func checkSentryEvent(t *testing.T, event *cue.Event, expected string) {
	req, err := getSentryCollector().formatRequest(event)
	if err != nil {
		t.Errorf("Encountered unexpected error formatting http request: %s", err)
	}
	requestJSON := cuetest.ParseRequestJSON(req)
	expectedJSON := cuetest.ParseStringJSON(expected)

	if cuetest.NestedFetch(requestJSON, "event_id") == "!(MISSING)" {
		t.Error("event_id is missing from request")
	}
	if cuetest.NestedFetch(requestJSON, "server_name") == "!(MISSING)" {
		t.Error("server_name is missing from request")
	}
	if cuetest.NestedFetch(requestJSON, "timestamp") == "!(MISSING)" {
		t.Error("timestamp is missing from request")
	}

	cuetest.NestedDelete(requestJSON, "event_id")
	cuetest.NestedDelete(expectedJSON, "event_id")
	cuetest.NestedDelete(requestJSON, "server_name")
	cuetest.NestedDelete(expectedJSON, "server_name")
	cuetest.NestedDelete(requestJSON, "timestamp")
	cuetest.NestedDelete(expectedJSON, "timestamp")
	cuetest.NestedCompare(t, requestJSON, expectedJSON)
}

func getSentryCollector() *sentryCollector {
	c := Sentry{
		DSN:          "https://public:private@app.getsentry.com.bogus/12345",
		ExtraContext: cue.NewContext("extra").WithValue("extra", "extra value"),
	}.New()
	sc, ok := c.(*sentryCollector)
	if !ok {
		panic(fmt.Sprintf("Expected to see a *sentryCollector but got %s instead", reflect.TypeOf(c)))
	}
	return sc
}

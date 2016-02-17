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

const honeybadgerJSON = `
{
  "error": {
    "backtrace": [
      {
        "file": "/path/github.com/bobziuchkovski/cue/frame3/file3.go",
        "method": "github.com/bobziuchkovski/cue/frame3.function3",
        "number": 3
      },
      {
        "file": "/path/github.com/bobziuchkovski/cue/frame2/file2.go",
        "method": "github.com/bobziuchkovski/cue/frame2.function2",
        "number": 2
      },
      {
        "file": "/path/github.com/bobziuchkovski/cue/frame1/file1.go",
        "method": "github.com/bobziuchkovski/cue/frame1.function1",
        "number": 1
      }
    ],
    "class": "errors.errorString",
    "message": "error event: error message",
    "tags": [
      "tag1",
      "tag2"
    ]
  },
  "notifier": {
    "name": "github.com/bobziuchkovski/cue",
    "url": "https://github.com/bobziuchkovski/cue",
    "version": "0.7.0"
  },
  "request": {
    "component": "github.com/bobziuchkovski/cue/frame3",
    "context": {
      "extra": "extra value",
      "k1": "some value",
      "k2": 2,
      "k3": 3.5,
      "k4": true
    }
  },
  "server": {
    "environment_name": "test",
    "hostname": "pegasus.bobbyz.org"
  }
}
`

const honeybadgerNoFramesJSON = `
{
  "error": {
    "class": "errors.errorString",
    "message": "error event: error message",
    "tags": [
      "tag1",
      "tag2"
    ]
  },
  "notifier": {
    "name": "github.com/bobziuchkovski/cue",
    "url": "https://github.com/bobziuchkovski/cue",
    "version": "0.7.0"
  },
  "request": {
    "context": {
      "extra": "extra value",
      "k1": "some value",
      "k2": 2,
      "k3": 3.5,
      "k4": true
    }
  },
  "server": {
    "environment_name": "test",
    "hostname": "pegasus.bobbyz.org"
  }
}
`

func TestHoneybadgerNilCollector(t *testing.T) {
	c := Honeybadger{}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the API key is missing, but got %s instead", c)
	}
}

func TestHoneybadger(t *testing.T) {
	checkHoneybadgerEvent(t, cuetest.ErrorEvent, honeybadgerJSON)
}

func TestHoneybadgerNoFrames(t *testing.T) {
	checkHoneybadgerEvent(t, cuetest.ErrorEventNoFrames, honeybadgerNoFramesJSON)
}

func TestHoneybadgerString(t *testing.T) {
	_ = fmt.Sprint(getHoneybadgerCollector())
}

func checkHoneybadgerEvent(t *testing.T, event *cue.Event, expected string) {
	req, err := getHoneybadgerCollector().formatRequest(event)
	if err != nil {
		t.Errorf("Encountered unexpected error formatting http request: %s", err)
	}
	requestJSON := cuetest.ParseRequestJSON(req)
	expectedJSON := cuetest.ParseStringJSON(expected)

	version := cuetest.NestedFetch(requestJSON, "notifier", "version")
	if version != fmt.Sprintf("%d.%d.%d", cue.Version.Major, cue.Version.Minor, cue.Version.Patch) {
		t.Errorf("Invalid notifier version: %s", version)
	}
	if cuetest.NestedFetch(requestJSON, "server", "hostname") == "!(MISSING)" {
		t.Error("Hostname is missing from request")
	}

	cuetest.NestedDelete(requestJSON, "notifier", "version")
	cuetest.NestedDelete(requestJSON, "server", "hostname")
	cuetest.NestedDelete(expectedJSON, "notifier", "version")
	cuetest.NestedDelete(expectedJSON, "server", "hostname")
	cuetest.NestedCompare(t, requestJSON, expectedJSON)
}

func getHoneybadgerCollector() *honeybadgerCollector {
	c := Honeybadger{
		Key:          "test",
		Tags:         []string{"tag1", "tag2"},
		Environment:  "test",
		ExtraContext: cue.NewContext("extra").WithValue("extra", "extra value"),
	}.New()
	hc, ok := c.(*honeybadgerCollector)
	if !ok {
		panic(fmt.Sprintf("Expected to see a *honeybadgerCollector but got %s instead", reflect.TypeOf(c)))
	}
	return hc
}

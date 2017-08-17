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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/format"
	"github.com/bobziuchkovski/cue/internal/cuetest"
)

func TestHTTPNilCollector(t *testing.T) {
	c := HTTP{}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the http request formatter is missing, but got %s instead", c)
	}
}

func TestHTTP(t *testing.T) {
	recorder := cuetest.NewHTTPRequestRecorder()
	s := httptest.NewServer(recorder)
	defer s.Close()

	c := HTTP{RequestFormatter: newHTTPRequestFormatter(s.URL)}.New()
	err := c.Collect(cuetest.DebugEvent)
	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}

	if len(recorder.Requests()) != 1 {
		t.Errorf("Expected exactly 1 request to be sent but saw %d instead", len(recorder.Requests()))
	}
	checkHTTPRequest(t, recorder.Requests()[0])
}

func TestHTTPError(t *testing.T) {
	recorder := cuetest.NewHTTPRequestRecorder()
	s := httptest.NewServer(recorder)
	defer s.Close()

	c := HTTP{
		RequestFormatter: newHTTPRequestFormatter(s.URL),
		Client:           &http.Client{Transport: cuetest.NewFailingHTTPTransport(1)},
	}.New()

	err := c.Collect(cuetest.DebugEvent)
	if err == nil {
		t.Error("Expected initial http request to fail, but it didn't")
	}
	err = c.Collect(cuetest.DebugEvent)
	if err != nil {
		t.Errorf("Encountered unexpected failure: %s", err)
	}

	if len(recorder.Requests()) != 1 {
		t.Errorf("Expected exactly 1 request to be sent but saw %d instead", len(recorder.Requests()))
	}
	checkHTTPRequest(t, recorder.Requests()[0])
}

func TestHTTP4XXErrorCode(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "test 400 error", 400)
	}))
	defer s.Close()

	c := HTTP{RequestFormatter: newHTTPRequestFormatter(s.URL)}.New()
	err := c.Collect(cuetest.DebugEvent)
	if err == nil {
		t.Error("Expected error but didn't receive one")
	}
}

func TestHTTP5XXErrorCode(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Error(w, "test 500 error", 500)
	}))
	defer s.Close()

	c := HTTP{RequestFormatter: newHTTPRequestFormatter(s.URL)}.New()
	err := c.Collect(cuetest.DebugEvent)
	if err == nil {
		t.Error("Expected error but didn't receive one")
	}
}

func TestHTTPStirng(t *testing.T) {
	c := HTTP{RequestFormatter: newHTTPRequestFormatter("http://bogus.private")}.New()

	// Ensure nothing panics
	_ = fmt.Sprint(c)
}

func checkHTTPRequest(t *testing.T, req *http.Request) {
	if req.Method != "POST" {
		t.Errorf("Expected POST method but saw %s instead", req.Method)
	}

	agentExpectation := fmt.Sprintf("github.com/bobziuchkovski/cue %d.%d.%d", cue.Version.Major, cue.Version.Minor, cue.Version.Patch)
	if req.Header.Get("User-Agent") != agentExpectation {
		t.Errorf("Expected User-Agent header of %q but saw %q instead", agentExpectation, req.Header.Get("User-Agent"))
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Errorf("Encountered unexpected error reading request body: %s", err)
	}

	bodyExpectation := "Jan  2 15:04:00 DEBUG file3.go:3 debug event k1=\"some value\" k2=2 k3=3.5 k4=true"
	if string(body) != bodyExpectation {
		t.Errorf("Expected to receive %q for request body but saw %q instead", bodyExpectation, string(body))
	}
}

func newHTTPRequestFormatter(url string) func(event *cue.Event) (*http.Request, error) {
	return func(event *cue.Event) (*http.Request, error) {
		return http.NewRequest("POST", url, strings.NewReader(format.RenderString(format.HumanReadable, event)))
	}
}

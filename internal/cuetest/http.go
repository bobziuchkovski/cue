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

package cuetest

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"sync"
)

// HTTPRequestRecorder implements http.RoundTripper, capturing all requests
// that are sent to it.
type HTTPRequestRecorder struct {
	mu       sync.Mutex
	requests []*http.Request
}

// NewHTTPRequestRecorder returns a new HTTPRequestRecorder instance.
func NewHTTPRequestRecorder() *HTTPRequestRecorder {
	return &HTTPRequestRecorder{}
}

// ServeHTTP is implemented to satisfy the http.RoundTripper interface.
func (rr *HTTPRequestRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(dump)
	dupe, err := http.ReadRequest(bufio.NewReader(buf))
	if err != nil {
		panic(err)
	}

	buf.Reset()
	buf.ReadFrom(req.Body)
	dupe.Body = ioutil.NopCloser(buf)

	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.requests = append(rr.requests, dupe)
}

// Requests returns a slice of the requests captured by the recorder.
func (rr *HTTPRequestRecorder) Requests() []*http.Request {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	return rr.requests
}

type failingHTTPTransport struct {
	succeedAfter int
	failCount    int
}

// NewFailingHTTPTransport returns a http.RoundTripper that fails requests
// until succeedAfter count have been submitted.  Afterwards, it passes
// requests to http.DefaultTransport.
func NewFailingHTTPTransport(succeedAfter int) http.RoundTripper {
	return &failingHTTPTransport{
		succeedAfter: succeedAfter,
	}
}

func (t *failingHTTPTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if t.failCount < t.succeedAfter {
		t.failCount++
		err = fmt.Errorf("%d more failures before I pass the HTTP request", t.succeedAfter-t.failCount)
		return
	}
	return http.DefaultTransport.RoundTrip(req)
}

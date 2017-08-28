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
	"net/http"

	"github.com/remerge/cue"
)

// HTTP represents configuration for http-based Collector instances. For each
// event, the collector calls RequestFormatter to generate a new http request.
// It then submits the request, setting a cue-specific User-Agent header.  The
// response status code is checked, but the content is otherwise ignored.  The
// collector treats 4XX and 5XX status codes as errors.
type HTTP struct {
	// Required
	RequestFormatter func(event *cue.Event) (*http.Request, error)

	// If specified, submit the generated requests via Client
	Client *http.Client
}

// New returns a new collector based on the HTTP configuration.
func (h HTTP) New() cue.Collector {
	if h.RequestFormatter == nil {
		log.Warn("HTTP.New called to created a collector, but RequestFormatter param is empty.  Returning nil collector.")
		return nil
	}
	if h.Client == nil {
		h.Client = &http.Client{}
	}
	return &httpCollector{HTTP: h}
}

func (h *httpCollector) String() string {
	return "HTTP(unknown, please wrap the HTTP collector and implement String())"
}

type httpCollector struct {
	HTTP
}

func (h *httpCollector) Collect(event *cue.Event) error {
	request, err := h.RequestFormatter(event)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", fmt.Sprintf("github.com/remerge/cue %d.%d.%d", cue.Version.Major, cue.Version.Minor, cue.Version.Patch))
	resp, err := h.Client.Do(request)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("cue/collector: http error: url=%s, error=%q", request.URL, err.Error())
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("cue/collector: http error: url=%s, code=%d", request.URL, resp.StatusCode)
	}
	return nil
}

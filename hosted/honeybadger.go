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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/collector"
	"github.com/bobziuchkovski/cue/format"
	"net/http"
)

// Honeybadger represents configuration for the Honeybadger service.  Collected
// events are sent to Honeybadger as new error occurrences, complete with
// relevant stack trace.  Honeybadger only supports error/fatal events, so
// collectors for the service should only be registered at the ERROR or FATAL
// log levels.
type Honeybadger struct {
	// Required
	Key string // Honeybadger API key

	// Optional
	Tags         []string    // Tags to send with every event
	ExtraContext cue.Context // Additional context values to send with every event
	Environment  string      // Environment name ("development", "production", etc.)
}

// New returns a new collector based on the Honeybadger configuration.
func (h Honeybadger) New() cue.Collector {
	if h.Key == "" {
		log.Warn("Honeybadger.New called to created a collector, but Key param is empty.  Returning nil collector.")
		return nil
	}
	return &honeybadgerCollector{
		Honeybadger: h,
		http:        collector.HTTP{RequestFormatter: h.formatRequest}.New(),
	}
}

func (h Honeybadger) formatRequest(event *cue.Event) (request *http.Request, err error) {
	body := format.RenderBytes(h.formatBody, event)
	request, err = http.NewRequest("POST", "https://api.honeybadger.io/v1/notices", bytes.NewReader(body))
	if err != nil {
		return
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", h.Key)
	return
}

func (h Honeybadger) formatBody(buffer format.Buffer, event *cue.Event) {
	post := &honeybadgerPost{
		Error:   h.errorFor(event),
		Request: h.requestFor(event),
		Server:  h.server(),
		Notifier: honeybadgerNotifier{
			Name:    "github.com/bobziuchkovski/cue",
			URL:     "https://github.com/bobziuchkovski/cue",
			Version: fmt.Sprintf("%d.%d.%d", cue.Version.Major, cue.Version.Minor, cue.Version.Patch),
		},
	}
	marshalled, _ := json.Marshal(post)
	buffer.Write(marshalled)
}

func (h Honeybadger) requestFor(event *cue.Event) honeybadgerRequest {
	pkg := ""
	if len(event.Frames) > 0 && event.Frames[0].Package != cue.UnknownPackage {
		pkg = event.Frames[0].Package
	}
	return honeybadgerRequest{
		Context:   cue.JoinContext("", event.Context, h.ExtraContext).Fields(),
		Component: pkg,
	}
}

func (h Honeybadger) errorFor(event *cue.Event) honeybadgerError {
	return honeybadgerError{
		Class:     format.RenderString(format.ErrorType, event),
		Message:   format.RenderString(format.MessageWithError, event),
		Tags:      h.Tags,
		Backtrace: h.backtraceFor(event),
	}
}

func (h Honeybadger) backtraceFor(event *cue.Event) []*honeybadgerFrame {
	var backtrace []*honeybadgerFrame
	for _, frame := range event.Frames {
		backtrace = append(backtrace, &honeybadgerFrame{
			Number: frame.Line,
			File:   frame.File,
			Method: frame.Function,
		})
	}
	return backtrace
}

func (h Honeybadger) server() honeybadgerServer {
	return honeybadgerServer{
		EnvironmentName: h.Environment,
		Hostname:        format.RenderString(format.FQDN, nil),
	}
}

type honeybadgerCollector struct {
	Honeybadger
	http cue.Collector
}

func (h *honeybadgerCollector) String() string {
	return fmt.Sprintf("Honeybadger(environment=%q)", h.Environment)
}

func (h *honeybadgerCollector) Collect(event *cue.Event) error {
	return h.http.Collect(event)
}

type honeybadgerPost struct {
	Notifier honeybadgerNotifier `json:"notifier"`
	Error    honeybadgerError    `json:"error"`
	Request  honeybadgerRequest  `json:"request"`
	Server   honeybadgerServer   `json:"server"`
}

type honeybadgerError struct {
	Class     string              `json:"class"`
	Message   string              `json:"message"`
	Tags      []string            `json:"tags,omitempty"`
	Backtrace []*honeybadgerFrame `json:"backtrace,omitempty"`
}

type honeybadgerFrame struct {
	Number int    `json:"number"`
	File   string `json:"file"`
	Method string `json:"method"`
}

type honeybadgerRequest struct {
	Context   cue.Fields `json:"context"`
	Component string     `json:"component,omitempty"`
}

type honeybadgerNotifier struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

type honeybadgerServer struct {
	EnvironmentName string `json:"environment_name,omitempty"`
	Hostname        string `json:"hostname"`
}

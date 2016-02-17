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
	"runtime"
)

// Rollbar represents configuration for the Rollbar service. Collected events
// are sent to Rollbar at matching event levels (debug, info, etc.), complete
// with relevant stack trace.
type Rollbar struct {
	// Required
	Token       string // Auth token
	Environment string // Environment name ("development", "production", etc.)

	// Optional
	ExtraContext     cue.Context // Additional context values to send with every event
	ProjectVersion   string      // Project version (SHA value, semantic version, etc.)
	ProjectFramework string      // Project framework name
}

// New returns a new collector based on the Rollbar configuration.
func (r Rollbar) New() cue.Collector {
	if r.Token == "" || r.Environment == "" {
		log.Warn("Rollbar.New called to created a collector, but Token or Environment param is empty.  Returning nil collector.")
		return nil
	}
	return &rollbarCollector{
		Rollbar: r,
		http:    collector.HTTP{RequestFormatter: r.formatRequest}.New(),
	}
}

func (r Rollbar) formatRequest(event *cue.Event) (request *http.Request, err error) {
	body := format.RenderBytes(r.formatBody, event)
	request, err = http.NewRequest("POST", "https://api.rollbar.com/api/1/item/", bytes.NewReader(body))
	if err != nil {
		return
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	return
}

func (r Rollbar) formatBody(buffer format.Buffer, event *cue.Event) {
	codever := r.ProjectVersion
	if len(codever) > 40 {
		codever = codever[:40]
	}

	bodyFormatter := r.formatTrace
	if event.Level > cue.ERROR || len(event.Frames) == 0 {
		bodyFormatter = r.formatMessage
	}

	contextJSON, _ := json.Marshal(cue.JoinContext("", event.Context, r.ExtraContext).Fields())
	marshalled, _ := json.Marshal(&rollbarPost{
		Token: r.Token,
		Data: rollbarData{
			Timestamp:   event.Time.Unix(),
			Environment: r.Environment,
			Framework:   r.ProjectFramework,
			Level:       rollbarLevel(event.Level),
			Body:        bodyFormatter(event),
			Custom:      json.RawMessage(contextJSON),
			CodeVersion: codever,
			Platform:    runtime.GOOS,
			Server: rollbarServer{
				Host: format.RenderString(format.FQDN, event),
			},
			Notifier: rollbarNotifier{
				Name:    "github.com/bobziuchkovski/cue",
				Version: fmt.Sprintf("%d.%d.%d", cue.Version.Major, cue.Version.Minor, cue.Version.Patch),
			},
			Language: "go",
		},
	})
	buffer.Write(marshalled)
}

func (r Rollbar) formatMessage(event *cue.Event) json.RawMessage {
	marshalled, _ := json.Marshal(&rollbarMessage{
		Message: &rollbarMessageBody{
			Body: format.RenderString(format.MessageWithError, event),
		},
	})
	return json.RawMessage(marshalled)
}

func (r Rollbar) formatTrace(event *cue.Event) json.RawMessage {
	body := &rollbarTraceBody{
		Trace: rollbarTrace{
			Exception: rollbarException{
				Class:       format.RenderString(format.ErrorType, event),
				Message:     event.Error.Error(),
				Description: format.RenderString(format.MessageWithError, event),
			},
		},
	}
	for i := len(event.Frames) - 1; i >= 0; i-- {
		body.Trace.Frames = append(body.Trace.Frames, &rollbarFrame{
			Filename: event.Frames[i].File,
			Lineno:   event.Frames[i].Line,
			Method:   event.Frames[i].Function,
		})
	}

	marshalled, _ := json.Marshal(body)
	return json.RawMessage(marshalled)
}

type rollbarCollector struct {
	Rollbar
	http cue.Collector
}

func (r *rollbarCollector) String() string {
	return fmt.Sprintf("Rollbar(environment=%s)", r.Environment)
}

func (r *rollbarCollector) Collect(event *cue.Event) error {
	return r.http.Collect(event)
}

type rollbarPost struct {
	Token string      `json:"access_token"`
	Data  rollbarData `json:"data"`
}

type rollbarData struct {
	Environment string          `json:"environment"`
	Body        json.RawMessage `json:"body"`
	Level       string          `json:"level"`
	Timestamp   int64           `json:"timestamp"`
	CodeVersion string          `json:"code_version,omitempty"`
	Platform    string          `json:"platform"`
	Language    string          `json:"language"`
	Framework   string          `json:"framework,omitempty"`
	Server      rollbarServer   `json:"server"`
	Custom      json.RawMessage `json:"custom"`
	Notifier    rollbarNotifier `json:"notifier"`
}

type rollbarServer struct {
	Host string `json:"host"`
}

type rollbarNotifier struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type rollbarMessage struct {
	Message *rollbarMessageBody `json:"message"`
}

type rollbarMessageBody struct {
	Body string `json:"body"`
}

type rollbarTraceBody struct {
	Trace rollbarTrace `json:"trace"`
}

type rollbarTrace struct {
	Frames    []*rollbarFrame  `json:"frames"`
	Exception rollbarException `json:"exception"`
}

type rollbarFrame struct {
	Filename string `json:"filename"`
	Lineno   int    `json:"lineno"`
	Method   string `json:"method"`
}

type rollbarException struct {
	Class       string `json:"class"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

func rollbarLevel(level cue.Level) string {
	switch level {
	case cue.DEBUG:
		return "debug"
	case cue.INFO:
		return "info"
	case cue.WARN:
		return "warning"
	case cue.ERROR:
		return "error"
	case cue.FATAL:
		return "critical"
	default:
		panic("cue/hosted: BUG invalid cue level")
	}
}

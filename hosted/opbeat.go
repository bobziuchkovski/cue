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

// Opbeat represents configuration for the Opbeat service.  Collected events
// are sent to Opbeat at matching event levels (debug, info, etc.), complete
// with relevant stack trace.
type Opbeat struct {
	// Required
	Token          string // Auth token
	AppID          string // Application ID
	OrganizationID string // Organization ID

	// Optional
	ExtraContext cue.Context // Additional context values to send with every event
}

// New returns a new collector based on the Opbeat configuration.
func (o Opbeat) New() cue.Collector {
	if o.Token == "" || o.AppID == "" || o.OrganizationID == "" {
		log.Warn("Opbeat.New called to created a collector, but Token, AppID, or OrganizationID param is empty.  Returning nil collector.")
		return nil
	}
	return &opbeatCollector{
		Opbeat: o,
		http:   collector.HTTP{RequestFormatter: o.formatRequest}.New(),
	}
}

func (o Opbeat) formatRequest(event *cue.Event) (request *http.Request, err error) {
	body := format.RenderBytes(o.formatBody, event)
	url := fmt.Sprintf("https://intake.opbeat.com/api/v1/organizations/%s/apps/%s/errors/", o.OrganizationID, o.AppID)
	request, err = http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+o.Token)
	return
}

func (o Opbeat) formatBody(buffer format.Buffer, event *cue.Event) {
	post := &opbeatPost{
		Timestamp:  event.Time.UTC().Format("2006-01-02T15:04:05.000Z"),
		Level:      opbeatLevel(event.Level),
		Logger:     event.Context.Name(),
		Message:    format.RenderString(format.MessageWithError, event),
		Culprit:    o.culpritFor(event),
		Extra:      cue.JoinContext("", event.Context, o.ExtraContext).Fields(),
		Exception:  o.exceptionFor(event),
		Stacktrace: o.stacktraceFor(event),
		Machine: opbeatMachine{
			Hostname: format.RenderString(format.FQDN, event),
		},
	}
	marshalled, _ := json.Marshal(post)
	buffer.Append(marshalled)
}

func (o Opbeat) culpritFor(event *cue.Event) string {
	if len(event.Frames) == 0 || event.Frames[0].Function == cue.UnknownFunction {
		return ""
	}
	return event.Frames[0].Function
}

func (o Opbeat) exceptionFor(event *cue.Event) *opbeatException {
	var exception *opbeatException
	if event.Level == cue.ERROR || event.Level == cue.FATAL {
		exception = &opbeatException{
			Type:  format.RenderString(format.ErrorType, event),
			Value: event.Error.Error(),
		}
		if len(event.Frames) != 0 {
			exception.Module = event.Frames[0].Package
		}
	}
	return exception
}

func (o Opbeat) stacktraceFor(event *cue.Event) *opbeatStacktrace {
	if len(event.Frames) == 0 {
		return nil
	}

	stacktrace := &opbeatStacktrace{}
	for i := len(event.Frames) - 1; i >= 0; i-- {
		stacktrace.Frames = append(stacktrace.Frames, &opbeatFrame{
			Filename: event.Frames[i].File,
			Function: event.Frames[i].Function,
			Lineno:   event.Frames[i].Line,
		})
	}
	return stacktrace
}

type opbeatCollector struct {
	Opbeat
	http cue.Collector
}

func (o *opbeatCollector) String() string {
	return fmt.Sprintf("Opbeat(appId=%s)", o.AppID)
}

func (o *opbeatCollector) Collect(event *cue.Event) error {
	return o.http.Collect(event)
}

type opbeatPost struct {
	Timestamp  string            `json:"timestamp"`
	Level      string            `json:"level"`
	Logger     string            `json:"logger,omitempty"`
	Message    string            `json:"message"`
	Culprit    string            `json:"culprit,omitempty"`
	Machine    opbeatMachine     `json:"machine"`
	Extra      cue.Fields        `json:"extra"`
	Exception  *opbeatException  `json:"exception,omitempty"`
	Stacktrace *opbeatStacktrace `json:"stacktrace,omitempty"`
}

type opbeatException struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Module string `json:"module,omitempty"`
}

type opbeatStacktrace struct {
	Frames []*opbeatFrame `json:"frames"`
}

type opbeatFrame struct {
	Filename string `json:"filename"`
	Function string `json:"function"`
	Lineno   int    `json:"lineno"`
}

type opbeatMachine struct {
	Hostname string `json:"hostname"`
}

func opbeatLevel(level cue.Level) string {
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
		return "fatal"
	default:
		panic("cue/hosted: BUG invalid level")
	}
}

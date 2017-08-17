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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/collector"
	"github.com/bobziuchkovski/cue/format"
)

const sentryVersion = 7

// Sentry represents configuration for the Sentry service. Collected events are
// sent to Sentry at matching event levels (debug, info, etc.), complete with
// relevant stack trace.
type Sentry struct {
	// Required
	DSN string // DSN for the app (e.g. https://<public>:<private>@app.getsentry.com/<appid>)

	// Optional
	ExtraContext   cue.Context // Additional context values to send with every event
	ProjectVersion string      // Project version (SHA value, semantic version, etc.)
}

// New returns a new collector based on the Sentry configuration.
func (s Sentry) New() cue.Collector {
	if s.DSN == "" || !validDSN(s.DSN) {
		log.Warn("Sentry.New called to created a collector, but DSN param is empty or invalid.  Returning nil collector.")
		return nil
	}
	return &sentryCollector{
		Sentry: s,
		http:   collector.HTTP{RequestFormatter: s.formatRequest}.New(),
	}
}

func (s Sentry) formatRequest(event *cue.Event) (request *http.Request, err error) {
	u, _ := url.Parse(s.DSN)
	body := format.RenderBytes(s.formatBody, event)
	request, err = http.NewRequest("POST", fmt.Sprintf("%s://%s/api%s/store/", u.Scheme, u.Host, u.Path), bytes.NewReader(body))
	if err != nil {
		return
	}

	secret, _ := u.User.Password()
	request.Header.Set("X-Sentry-Auth", formatSentryAuth(u.User.Username(), secret))
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	return
}

func (s Sentry) formatBody(buffer format.Buffer, event *cue.Event) {
	// Sentry states their maximum supported message size is 1000 chars
	message := format.RenderString(format.MessageWithError, event)
	if len(message) > 1000 {
		message = message[:1000]
	}

	post := &sentryPost{
		Timestamp:  event.Time.UTC().Format("2006-01-02T15:04:05"),
		EventID:    hex.EncodeToString(uuid()),
		Message:    message,
		Exception:  s.exceptionFor(event),
		Culprit:    s.culpritFor(event),
		Tags:       s.tagsFor(event),
		Release:    s.ProjectVersion,
		Logger:     event.Context.Name(),
		Level:      sentryLevel(event.Level),
		ServerName: format.RenderString(format.FQDN, event),
		Platform:   "go",
	}

	marshalled, _ := json.Marshal(post)
	buffer.Append(marshalled)
}

func (s Sentry) exceptionFor(event *cue.Event) *sentryException {
	var exception *sentryException
	if event.Level == cue.ERROR || event.Level == cue.FATAL {
		exception = &sentryException{
			Type:       format.RenderString(format.ErrorType, event),
			Value:      event.Message,
			Stacktrace: s.stacktraceFor(event),
		}
		if len(event.Frames) > 0 && event.Frames[0].Package != cue.UnknownPackage {
			exception.Module = event.Frames[0].Package
		}
	}
	return exception
}

func (s Sentry) culpritFor(event *cue.Event) string {
	if len(event.Frames) == 0 || event.Frames[0].Function == cue.UnknownFunction {
		return ""
	}
	return event.Frames[0].Function
}

func (s Sentry) stacktraceFor(event *cue.Event) *sentryStacktrace {
	if len(event.Frames) == 0 {
		return nil
	}

	stacktrace := &sentryStacktrace{}
	for i := len(event.Frames) - 1; i >= 0; i-- {
		stacktrace.Frames = append(stacktrace.Frames, &sentryFrame{
			Filename: event.Frames[i].File,
			Function: event.Frames[i].Function,
			Module:   event.Frames[i].Package,
			Lineno:   event.Frames[i].Line,
		})
	}
	return stacktrace
}

func (s Sentry) tagsFor(event *cue.Event) []sentryTag {
	var tags []sentryTag
	cue.JoinContext("", event.Context, s.ExtraContext).Each(func(key string, value interface{}) {
		tags = append(tags, sentryTag{Name: key, Value: fmt.Sprint(value)})
	})
	return tags
}

func validDSN(dsn string) bool {
	u, err := url.Parse(dsn)
	if err != nil {
		return false
	}
	if u.User == nil {
		return false
	}
	if u.User.Username() == "" {
		return false
	}
	_, haspass := u.User.Password()
	if !haspass {
		return false
	}
	return true
}

type sentryCollector struct {
	Sentry
	appID string
	http  cue.Collector
}

func (s *sentryCollector) String() string {
	return fmt.Sprintf("Sentry(app=%s)", s.appID)
}

func (s *sentryCollector) Collect(event *cue.Event) error {
	return s.http.Collect(event)
}

type sentryPost struct {
	EventID   string `json:"event_id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Logger    string `json:"logger,omitempty"`
	Platform  string `json:"platform"`

	// For errors
	Exception *sentryException `json:"exception,omitempty"`

	// Optional attrs
	Culprit    string      `json:"culprit,omitempty"`
	ServerName string      `json:"server_name"`
	Release    string      `json:"release,omitempty"`
	Tags       []sentryTag `json:"tags,omitempty"`
}

type sentryException struct {
	Type       string            `json:"type"`
	Value      string            `json:"value"`
	Module     string            `json:"module,omitempty"`
	Stacktrace *sentryStacktrace `json:"stacktrace,omitempty"`
}

type sentryStacktrace struct {
	Frames []*sentryFrame `json:"frames"`
}

type sentryFrame struct {
	Filename string `json:"filename"`
	Function string `json:"function"`
	Module   string `json:"module"`
	Lineno   int    `json:"lineno"`
}

type sentryTag struct {
	Name  string
	Value string
}

func (tag sentryTag) MarshalJSON() ([]byte, error) {
	pair := []string{tag.Name, tag.Value}
	return json.Marshal(pair)
}

func formatSentryAuth(public, private string) string {
	auth := format.GetBuffer()
	defer format.ReleaseBuffer(auth)
	auth.AppendString(fmt.Sprintf("Sentry sentry_version=%d, ", sentryVersion))
	auth.AppendString(fmt.Sprintf("sentry_client=github.com/bobziuchkovski/cue:%d.%d.%d, ", cue.Version.Major, cue.Version.Minor, cue.Version.Patch))
	auth.AppendString(fmt.Sprintf("sentry_timestamp=%d, ", time.Now().UTC().Unix()))
	auth.AppendString(fmt.Sprintf("sentry_key=%s, ", public))
	auth.AppendString(fmt.Sprintf("sentry_secret=%s", private))
	return string(auth.Bytes())
}

func sentryLevel(level cue.Level) string {
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
		panic("cue/hosted: BUG invalid cue level")
	}
}

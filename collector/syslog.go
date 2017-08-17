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
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/format"
)

var (
	// RFC 5324 compliance requires this byte-order mark right before the message string
	rfc5424BOM = []byte{0xef, 0xbb, 0xbf}

	// Default syslog socket locations
	syslogSockets = []string{"/dev/log", "/var/run/log", "/var/run/syslog"}

	// Facility names
	facilityNames = map[Facility]string{
		KERN:     "KERN",
		USER:     "USER",
		MAIL:     "MAIL",
		DAEMON:   "DAEMON",
		AUTH:     "AUTH",
		SYSLOG:   "SYSLOG",
		LPR:      "LPR",
		NEWS:     "NEWS",
		UUCP:     "UUCP",
		CRON:     "CRON",
		AUTHPRIV: "AUTHPRIV",
		FTP:      "FTP",
		NTP:      "NTP",
		AUDIT:    "AUDIT",
		ALERT:    "ALERT",
		LOCAL0:   "LOCAL0",
		LOCAL1:   "LOCAL1",
		LOCAL2:   "LOCAL2",
		LOCAL3:   "LOCAL3",
		LOCAL4:   "LOCAL4",
		LOCAL5:   "LOCAL5",
		LOCAL6:   "LOCAL6",
		LOCAL7:   "LOCAL7",
	}
)

const (
	rfc5424Time    = "2006-01-02T15:04:05.000000-07:00"
	rfc5424Version = "1"
	ourID          = "cue@47338"
	syslogNil      = "-"
)

type priority uint

// Facility represents syslog facilities for the Syslog and StructuredSyslog
// collectors.
type Facility uint

// String returns a string representation of the facility name
func (f Facility) String() string {
	name, present := facilityNames[f]
	if present {
		return name
	}
	return "INVALID"
}

// Facility constants for the Syslog and StructuredSyslog collectors.
const (
	KERN Facility = iota
	USER
	MAIL
	DAEMON
	AUTH
	SYSLOG
	LPR
	NEWS
	UUCP
	CRON
	AUTHPRIV
	FTP
	NTP
	AUDIT
	ALERT
	_
	LOCAL0
	LOCAL1
	LOCAL2
	LOCAL3
	LOCAL4
	LOCAL5
	LOCAL6
	LOCAL7
)

type severity uint

const (
	sEMERGENCY severity = iota
	sALERT
	sCRITICAL
	sERROR
	sWARN
	sNOTICE
	sINFO
	sDEBUG
)

// Syslog represents configuration for traditional RFC 3339 (unstructured/BSD)
// syslog collector instances.
//
// The MessageFormatter must ensure new line characters in event messages are
// properly escaped.  The default formatter, format.HumanMessage, does this
// automatically.
type Syslog struct {
	App      string
	Facility Facility

	// Optional Socket config.  Defaults to a local unix socket.
	Network string
	Address string
	TLS     *tls.Config

	// Optional extras
	Formatter format.Formatter // Default: format.HumanMessage
}

// New returns a new collector based on the Syslog configuration.
func (s Syslog) New() cue.Collector {
	if s.App == "" {
		log.Warn("Syslog.New called to created a collector, but App param is empty.  Returning nil collector.")
		return nil
	}

	var err error
	if s.Network == "" || s.Address == "" {
		s.Network, s.Address, err = localSyslog()
	}
	if err != nil {
		log.Warn("Syslog.New called to created a collector, but Network or Address param is empty.  Couldn't find a local syslog socket either.  Returning nil collector.")
		return nil
	}

	local := false
	if s.Network == "unix" || s.Network == "unixgram" {
		local = true
	}

	return &syslogCollector{
		Syslog: s,
		socket: Socket{
			Formatter: syslogFormatter(s.Facility, s.App, local, s.Formatter),
			Network:   s.Network,
			Address:   s.Address,
			TLS:       s.TLS,
		}.New(),
	}
}

type syslogCollector struct {
	Syslog
	socket cue.Collector
}

func (s *syslogCollector) String() string {
	return fmt.Sprintf("Syslog(app=%s, facility=%s, network=%s, address=%s, tls=%t)", s.App, s.Facility, s.Network, s.Address, s.TLS != nil)
}

func (s *syslogCollector) Collect(event *cue.Event) error {
	return s.socket.Collect(event)
}

func (s *syslogCollector) Close() error {
	return s.socket.(io.Closer).Close()
}

func syslogFormatter(facility Facility, app string, local bool, msgFormatter format.Formatter) format.Formatter {
	if msgFormatter == nil {
		msgFormatter = format.HumanMessage
	}

	formatter := format.Formatf("%v%v %v %v: %v\n", priFormatter(facility), format.Time(time.RFC3339), format.Hostname, procIDFormatter(app), msgFormatter)
	if local {
		formatter = format.Formatf("%v%v %v: %v\n", priFormatter(facility), format.Time(time.Stamp), procIDFormatter(app), msgFormatter)
	}
	// RFC 3164 explicitly limits the message length to 1024 bytes
	return format.Truncate(formatter, 1024)
}

// StructuredSyslog represents configuration for RFC 5424 (structured) syslog
// Collector instances.  Messages are written with the appropriate header
// according to the provided facility and app params.  Context data is written
// as structured key=value pairs in the message header.
//
// The MessageFormatter must ensure new line characters in event messages are
// properly escaped.  The default formatter, format.HumanMessage, does this
// automatically.
//
// Please note that the default StructuredFormatter, format.StructuredContext,
// silently drops context key/value pairs if the key name doesn't match
// RFC 5424 requirements (longer than 32 chars, contains non-ASCII or
// control characters, etc.).
type StructuredSyslog struct {
	// Required
	Facility Facility
	App      string

	// Optional Socket config.  Defaults to a local unix socket if available.
	Network string
	Address string
	TLS     *tls.Config

	// Optional extras
	MessageFormatter    format.Formatter // Default: format.HumanMessage
	StructuredFormatter format.Formatter // Default: format.StructuredContext
	ID                  string           // Default: cue@47338

	// RFC5424 requires a byte-order mark (BOM) prior to the message text.
	// However, not all syslog servers expect or even understand it.
	WriteBOM bool
}

// New returns a new collector based on the StructuredSyslog configuration.
func (s StructuredSyslog) New() cue.Collector {
	if s.App == "" {
		log.Warn("StructuredSyslog.New called to created a collector, but App param is empty.  Returning nil collector.")
		return nil
	}

	var err error
	if s.Network == "" || s.Address == "" {
		s.Network, s.Address, err = localSyslog()
	}
	if err != nil {
		log.Warn("StructuredSyslog.New called to created a collector, but Network or Address param is empty.  Couldn't find a local syslog socket either.  Returning nil collector.")
		return nil
	}

	return &structuredCollector{
		StructuredSyslog: s,
		socket: Socket{
			Formatter: structuredFormatter(s.Facility, s.App, s.MessageFormatter, s.StructuredFormatter, s.ID, s.WriteBOM),
			Network:   s.Network,
			Address:   s.Address,
			TLS:       s.TLS,
		}.New(),
	}
}

type structuredCollector struct {
	StructuredSyslog
	socket cue.Collector
}

func (s *structuredCollector) String() string {
	return fmt.Sprintf("StructuredSyslog(app=%s, facility=%s, network=%s, address=%s, tls=%t)", s.App, s.Facility, s.Network, s.Address, s.TLS != nil)
}

func (s *structuredCollector) Collect(event *cue.Event) error {
	return s.socket.Collect(event)
}

func (s *structuredCollector) Close() error {
	return s.socket.(io.Closer).Close()
}

func structuredFormatter(facility Facility, app string, msgFormatter format.Formatter, structFormatter format.Formatter, ID string, writeBom bool) format.Formatter {
	msgid := syslogNil
	bomFormatter := format.Literal("")
	if writeBom {
		bomFormatter = formatBOM
	}
	if ID == "" {
		ID = ourID
	}
	if msgFormatter == nil {
		msgFormatter = format.HumanMessage
	}
	if structFormatter == nil {
		structFormatter = format.StructuredContext
	}
	return format.Formatf("%v%v %v %v %v %v %v [%v] %v%v\n",
		priFormatter(facility), format.Literal(rfc5424Version), format.Time(rfc5424Time),
		format.FQDN, format.Literal(app), procIDFormatter(app), format.Literal(msgid),
		format.Join(" ", format.Literal(ID), structFormatter), bomFormatter, msgFormatter)
}

func localSyslog() (network string, address string, err error) {
	for _, network = range []string{"unixgram", "unix"} {
		for _, address = range syslogSockets {
			_, err = net.Dial(network, address)
			if err == nil {
				return
			}
		}
	}
	err = errors.New("cue/collector: failed to find unix socket for syslog")
	return
}

func formatBOM(buf format.Buffer, event *cue.Event) {
	buf.Append(rfc5424BOM)
}

func priFormatter(facility Facility) format.Formatter {
	return func(buf format.Buffer, event *cue.Event) {
		buf.AppendString(fmt.Sprintf("<%d>", priorityFor(facility, event.Level)))
	}
}

func procIDFormatter(app string) format.Formatter {
	return format.Literal(fmt.Sprintf("%s[%d]", app, os.Getpid()))
}

func priorityFor(facility Facility, level cue.Level) priority {
	return priority(8*facility) + priority(severityFor(level))
}

func severityFor(level cue.Level) severity {
	switch level {
	case cue.DEBUG:
		return sDEBUG
	case cue.INFO:
		return sINFO
	case cue.WARN:
		return sWARN
	case cue.ERROR:
		return sERROR
	case cue.FATAL:
		return sCRITICAL
	default:
		panic(fmt.Errorf("cue/collector: unknown level: %s", level))
	}
}

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
	"crypto/tls"
	"fmt"
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/collector"
	"github.com/bobziuchkovski/cue/format"
	"io"
)

const logglyNetwork = "tcp"
const logglyAddress = "logs-01.loggly.com:514"
const logglyEnterpriseID = 41058

// Loggly represents configuration for the Loggly service.
//
// By default, logs are transported -in clear text-.  This is very bad for
// security.  Please see the example for enabling TLS transport encryption with
// Loggly.
type Loggly struct {
	// Required
	Token    string             // Loggly auth token.  Omit the trailing @41058 from this value.
	App      string             // Syslog app name
	Facility collector.Facility // Syslog facility

	// Optional socket config
	Network string      // Default: "tcp"
	Address string      // Default: "logs-01.loggly.com:514"
	TLS     *tls.Config // TLS transport config

	// Optional formatting config
	Formatter format.Formatter // Default: format.JSONMessage
	Tags      []string         // Tags to send with every event
}

// New returns a new collector based on the Loggly configuration.
func (l Loggly) New() cue.Collector {
	if l.Token == "" {
		log.Warn("Loggly.New called to created a collector, but the Token param is empty.  Returning nil collector.")
		return nil
	}

	if l.App == "" {
		log.Warn("Loggly.New called to created a collector, but the App param is empty.  Returning nil collector.")
		return nil
	}

	if l.Network == "" {
		l.Network = logglyNetwork
	}
	if l.Address == "" {
		l.Address = logglyAddress
	}
	if l.Formatter == nil {
		l.Formatter = format.JSONMessage
	}

	return &logglyCollector{
		Loggly: l,
		syslog: collector.StructuredSyslog{
			Facility:            l.Facility,
			App:                 l.App,
			Network:             l.Network,
			Address:             l.Address,
			TLS:                 l.TLS,
			MessageFormatter:    l.Formatter,
			StructuredFormatter: l.structuredFormatter(),
			ID:                  fmt.Sprintf("%s@%d", l.Token, logglyEnterpriseID),
		}.New(),
	}
}

func (l Loggly) structuredFormatter() format.Formatter {
	var literals []format.Formatter
	for _, tag := range l.Tags {
		literals = append(literals, format.Literal(fmt.Sprintf("tag=%q", tag)))
	}
	return format.Join(" ", literals...)
}

type logglyCollector struct {
	Loggly
	syslog cue.Collector
}

func (l *logglyCollector) String() string {
	return fmt.Sprintf("Loggly(app=%s, facility=%s, tls=%t)", l.App, l.Facility, l.TLS != nil)
}

func (l *logglyCollector) Collect(event *cue.Event) error {
	return l.syslog.Collect(event)
}

func (l *logglyCollector) Close() error {
	return l.syslog.(io.Closer).Close()
}

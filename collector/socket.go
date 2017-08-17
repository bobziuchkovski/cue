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
	"fmt"
	"net"

	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/format"
)

// Socket represents configuration for socket-based Collector instances. The
// collector writes messages to a connection specified by the network, address,
// and (optionally) TLS params.  The socket connection is opened via net.Dial,
// or by tls.Dial if TLS config is specified.  See the net and crypto/tls
// packages for details on supported Network and Address specifications.
type Socket struct {
	// Required
	Network string
	Address string

	// Optional
	TLS       *tls.Config
	Formatter format.Formatter // Default: format.HumanReadable
}

// New returns a new collector based on the Socket configuration.
func (s Socket) New() cue.Collector {
	if s.Network == "" {
		log.Warn("Socket.New called to created a collector, but Network param is empty.  Returning nil collector.")
		return nil
	}
	if s.Address == "" {
		log.Warn("Socket.New called to created a collector, but Address param is empty.  Returning nil collector.")
		return nil
	}
	if s.Formatter == nil {
		s.Formatter = format.HumanReadable
	}
	return &socketCollector{Socket: s}
}

type socketCollector struct {
	Socket
	conn      net.Conn
	connected bool
}

func (s *socketCollector) String() string {
	return fmt.Sprintf("Socket(network=%s, address=%s, tls=%t)", s.Network, s.Address, s.TLS != nil)
}

func (s *socketCollector) Collect(event *cue.Event) error {
	if !s.connected {
		err := s.reopen()
		if err != nil {
			return err
		}
	}

	buf := format.GetBuffer()
	defer format.ReleaseBuffer(buf)
	s.Formatter(buf, event)

	_, err := s.conn.Write(buf.Bytes())
	if err != nil {
		s.conn.Close()
		s.conn = nil
		s.connected = false
	}
	return err
}

func (s *socketCollector) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *socketCollector) reopen() error {
	var err error
	if s.TLS != nil {
		s.conn, err = tls.Dial(s.Network, s.Address, s.TLS)
		if err == nil {
			s.connected = true
		}
		return err
	}
	s.conn, err = net.Dial(s.Network, s.Address)
	if err == nil {
		s.connected = true
	}
	return err
}

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
	"bytes"
	"crypto/tls"
	"net"
	"reflect"
	"sync"
	"testing"
)

// NetRecorder is an interface representing a network listener/recorder.  The
// recorder stores all content sent to it.  Recorders are created in an
// unstarted state and must be explicitly started via the Start() method and
// explicitly stopped via the Close() method once finished.
type NetRecorder interface {
	// Address returns the address string for the recorder.
	Address() string

	// Contents returns the bytes that have been sent to the recorder.
	Contents() []byte

	// CheckByteContents checks if the bytes captured by the recorder match the
	// given expectation.  If not, t.Errorf is called with a comparison.
	CheckByteContents(t *testing.T, expectation []byte)

	// CheckStringContents checks if the bytes captured by the recorder match
	// the given string expectation.  If not, t.Errorf is called with a
	// comparison.
	CheckStringContents(t *testing.T, expectation string)

	// Start starts the recorder.
	Start()

	// Close stops the recorder and terminates any active connections.
	Close() error

	// Done returns a channel that blocks until the recorder is finished.
	Done() <-chan struct{}

	// Err returns the first error encountered by the recorder, if any.
	Err() error
}

type netRecorder struct {
	done   chan struct{}
	cancel chan struct{}
	err    *firstError

	startOnce sync.Once
	closeOnce sync.Once

	mu        sync.Mutex
	network   string
	address   string
	enableTLS bool
	content   []byte
	listener  net.Listener
}

// NewTCPRecorder returns a NetRecorder that listens for TCP connections.
func NewTCPRecorder() NetRecorder {
	return newNetRecorder("tcp", randomAddress(), false)
}

// NewTLSRecorder returns a NetRecorder that listens for TCP connections using
// TLS transport encryption.
func NewTLSRecorder() NetRecorder {
	return newNetRecorder("tcp", randomAddress(), true)
}

func newNetRecorder(network, address string, enableTLS bool) NetRecorder {
	return &netRecorder{
		done:      make(chan struct{}),
		cancel:    make(chan struct{}),
		network:   network,
		address:   address,
		enableTLS: enableTLS,
		err:       &firstError{},
	}
}

func (nr *netRecorder) Start() {
	nr.startOnce.Do(func() {
		var err error
		var listener net.Listener

		if nr.enableTLS {
			cert, err := tls.LoadX509KeyPair("test.crt", "test.key")
			if err != nil {
				panic(err)
			}
			tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
			listener, err = tls.Listen(nr.network, nr.address, tlsConfig)
		} else {
			listener, err = net.Listen(nr.network, nr.address)
		}
		if err == nil {
			nr.listener = listener
			go nr.run()
		} else {
			panic(err)
		}
	})
}

func (nr *netRecorder) Address() string {
	return nr.address
}

func (nr *netRecorder) Contents() []byte {
	<-nr.done

	nr.mu.Lock()
	defer nr.mu.Unlock()
	return nr.content
}

func (nr *netRecorder) CheckByteContents(t *testing.T, expectation []byte) {
	if !reflect.DeepEqual(nr.Contents(), expectation) {
		t.Errorf("Expected recorded content of %x but got %x instead", expectation, nr.Contents())
	}
}

func (nr *netRecorder) CheckStringContents(t *testing.T, expectation string) {
	if string(nr.Contents()) != expectation {
		t.Errorf("Expected recorded content of %q but got %q instead", expectation, nr.Contents())
	}
}

func (nr *netRecorder) Done() <-chan struct{} {
	return nr.done
}

func (nr *netRecorder) Err() error {
	<-nr.done
	return nr.err.Error()
}

func (nr *netRecorder) Close() error {
	nr.closeOnce.Do(func() {
		close(nr.cancel)
		if nr.listener != nil {
			nr.listener.Close()
			<-nr.done
		}
	})

	return nr.err.Error()
}

func (nr *netRecorder) run() {
	conn, err := nr.listener.Accept()
	if err != nil {
		nr.err.Set(err)
		close(nr.done)
		return
	}

	go func(conn net.Conn) {
		<-nr.cancel
		nr.err.Set(conn.Close())
	}(conn)

	var buf bytes.Buffer
	_, err = buf.ReadFrom(conn)
	nr.err.Set(err)

	nr.mu.Lock()
	defer nr.mu.Unlock()
	nr.content = buf.Bytes()
	close(nr.done)
}

type firstError struct {
	mu  sync.Mutex
	err error
}

func (se *firstError) Error() error {
	se.mu.Lock()
	defer se.mu.Unlock()
	return se.err
}

func (se *firstError) Set(err error) {
	se.mu.Lock()
	defer se.mu.Unlock()
	if se.err == nil {
		se.err = err
	}
}

func randomAddress() string {
	l, err := net.Listen("tcp", "localhost:0")
	defer l.Close()

	if err != nil {
		panic(err)
	}
	return l.Addr().String()
}

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

/*
Package collector implements event collection.

Implementations

This package provides event collection to file, syslog, web servers, and
network sockets.

Nil Instances

Collector implementations emit a WARN log event and return a nil collector
instance if required parameters are missing.  The cue.Collect and
cue.CollectAsync functions treat nil collectors as a no-op, so this is
perfectly safe.

Implementing Custom Collectors

Implementing a new cue.Collector is easy.  The Collect method is the only
method in the interface, and cue ensures the method is only called by a
single goroutine.  No additional synchronization is required.  If the
the collector implements the io.Closer interface, it's Close method will be
called when terminated.  See the implementations in this package for examples.

Collector Failures and Degradation

Where possible, collector implementations attempt to recover from errors.
The File collector attempts closing and re-opening its file handle, and the
socket and syslog collectors close and open new network connections.  Thus
transient errors will recover automatically if the source of the problem is
resolved.  However, collectors must still return error values for visibility
and handling by cue workers.

If a collector returns an error, cue will re-send the event to the collector
2 additional times before giving up.  After the third try, cue puts the
collector into a degraded state and prevents it from collecting new events.
It then emits an error event to all other collectors to surface the
degradation.  Finally, it tries sending an error event to the degraded
collector indefinitely until it succeeds, using exponential backoff with a
maximum delay of 5 minutes between attempts.  If the collector successfully
sends the error event, the collector is marked healthy again and a WARN event
is emitted to notify other collectors of the returned health.

If a collector panics, cue recovers the panic, discards the collector, and
emits a FATAL event to other collectors for visibility.
*/
package collector

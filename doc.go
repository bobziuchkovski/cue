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
Package cue implements contextual logging with "batteries included".  It has
thorough test coverage and supports logging to stdout/stderr, file, syslog,
and network sockets, as well as hosted third-party logging and error/reporting
services such as Honeybadger, Loggly, Opbeat, Rollbar, and Sentry.

Cue uses atomic operations to compare logging calls to registered collector
thresholds.  This ensures no-op calls are performed quickly and without lock
contention.  On a 2015 MacBook Pro, no-op calls take about 16ns/call, meaning
tens of millions of calls may be dispatched per second.  Uncollected log calls
are very cheap.

Furthermore, collector thresholds may be altered dynamically at run-time, on a
per-collector basis.  If debugging logs are needed to troubleshoot a live issue,
collector thresholds may be set to the DEBUG level for a short period of time
and then restored to their original levels shortly thereafter.  See the SetLevel
function for details.

Basics

Logging instances are created via the NewLogger function.  A simple convention
is to initialize an unexported package logger:

	var log = cue.NewLogger("some/package/name")

Additional context information may be added to the package logger via the
log.WithValue and log.WithFields methods:

	func DoSomething(user string) {
		log.WithValue("user", user).Info("Doing something")
	}

	func DoSomethingElse(user string, authorized bool) {
		log.WithFields(cue.Fields{
			"user": user,
			"authorized": authorized,
		}).Info("Something else requested")
	}

Depending on the collector and log format, output would look something like:

	<DATE> INFO Something else requested user=<user> authorized=<authorized>

Error Logging and Recovery

Cue simplifies error reporting by logging the given error and message, and then
returning the same error value.  Hence you can return the log.Error/log.Errorf
values in-line:

	filename := "somefile"
	f, err := os.Create(filename)
	if err != nil {
		return log.Errorf(err, "Failed to create %q", filename)
	}

Cue provides Collector implementations for popular error reporting services
such as Honeybadger, Rollbar, Sentry, and Opbeat.  If one of these collector
implementations were registered, the above code would automatically open a new
error report, complete with stack trace and context information from the logger
instance.  See the cue/hosted package for details.

Finally, cue provides convenience methods for panic and recovery. Calling Panic
or Panicf will log the provided message at the FATAL level and then panic.
Calling Recover recovers from panics and logs the recovered value and message
at the FATAL level.

	func doSomething() {
		defer log.Recover("Recovered panic in doSomething")
		doSomethingThatPanics()
	}

If a panic is triggered via a cue logger instance's Panic or Panicf methods,
Recover recovers from the panic but only emits the single event from the
Panic/Panicf method.

Event Collection

Cue decouples event generation from event collection.  Library and framework
authors may generate log events without concern for the details of collection.
Event collection is opt-in -- no collectors are registered by default.

Event collection, if enabled, should be configured close to a program's main
package/function, not by libraries.  This gives the event subscriber complete
control over the behavior of event collection.

Collectors are registered via the Collect and CollectAsync functions.  Each
collector is registered for a given level threshold.  The threshold for a
collector may be updated at any time using the SetLevel function.

Collect registers fully synchronous event collectors.  Logging calls that match
a synchronous collector's threshold block until the collector's Collect method
returns successfully.  This is dangerous if the Collector performs any
operations that block or return errors.  However, it's simple to use and
understand:

	func main() {
		// Follow a 12-factor approach and log unbuffered to stdout.
		// See http://12factor.net for details.
		cue.Collect(cue.INFO, collector.Terminal{}.New())
		defer log.Recover("Recovered from panic in main")

		RunTheProgram()
	}

CollectAsync registers asynchronous collectors.  It creates a buffered channel
for the collector and starts a worker goroutine to service events.  Logging
calls return after queuing events to the collector channel.  If the channel's
buffer is full, the event is dropped and a drop counter is incremented
atomically.  This ensures asynchronous logging calls never block.  The worker
goroutine detects changes in the atomic drop counter and surfaces drop events
as collector errors.  See the cue/collector docs for details on collector
error handling.

When asynchronous logging is enabled, Close must be called to flush queued
events on program termination.  Close is safe to call even if asynchronous
logging isn't enabled -- it returns immediately if no events are queued.
Note that ctrl+c and kill <pid> terminate Go programs without triggering
cleanup code.  When using asynchronous logging, it's a good idea to register
signal handlers to capture SIGINT (ctrl+c) and SIGTERM (kill <pid>).  See the
os/signals package docs for details.

	func main() {
		// Use async logging to local syslog
		cue.CollectAsync(cue.INFO, 10000, collector.Syslog{
			App: "theapp",
			Facility: collector.LOCAL0,
		}.New())

		// Close/flush buffered events on program termination.
		// Note that this won't fire if ctrl+c is used or kill <pid>.  You need
		// to install signal handlers for SIGINT/SIGTERM to handle those cases.
		defer cue.Close(5 * time.Second)

		defer log.Recover("Recovered from panic in main")
		RunTheProgram()
	}

Stack Frame Collection

By default, cue collects a single stack frame for any event that matches a
registered collector.  This ensures collectors may log the file name, package,
and line number for any collected event.  SetFrames may be used to alter this
frame count, or disable frame collection entirely.  See the SetFrames function
for details.

When using error reporting services, SetFrames should be used to increase the
errorFrames parameter from the default value of 1 to a value that provides
enough stack context to successfully diagnose reported errors.
*/
package cue

[![Build Status](https://travis-ci.org/bobziuchkovski/cue.svg?branch=master)](https://travis-ci.org/bobziuchkovski/cue)
[![Coverage](https://gocover.io/_badge/github.com/bobziuchkovski/cue?1)](https://gocover.io/github.com/bobziuchkovski/cue)
[![Report Card](http://goreportcard.com/badge/bobziuchkovski/cue)](http://goreportcard.com/report/bobziuchkovski/cue)
[![GoDoc](https://godoc.org/github.com/bobziuchkovski/cue?status.svg)](https://godoc.org/github.com/bobziuchkovski/cue)

# Cue

## Overview

Cue implements contextual logging with "batteries included".  It has
thorough test coverage and supports logging to stdout/stderr, file, syslog,
and network sockets, as well as hosted third-party logging and error/reporting
services such as Honeybadger, Loggly, Opbeat, Rollbar, and Sentry.

Cue uses atomic operations to compare logging calls to registered collector
thresholds.  This ensures no-op calls are performed quickly and without lock
contention.  On a 2015 MacBook Pro, no-op calls take about 16ns/call, meaning
tens of millions of calls may be dispatched per second.  Uncollected log calls
are very cheap.

## API Promise

Minor breaking changes may occur prior to the 1.0 release.  After the 1.0
release, the API is guaranteed to remain backwards compatible.

_Cue makes use of sync/atomic.Value and thus requires Go 1.4.x or later._

## Basic Use

Logging instances are created via the NewLogger function.  A simple convention
is to initialize an unexported package logger:

```go
	var log = cue.NewLogger("some/package/name")
```

Additional context information may be added to the package logger via the
log.WithValue and log.WithFields methods:

```go
	func DoSomething(user string) {
		log.WithValue("user", user).Info("Doing something")
	}

	func DoSomethingElse(user string, authorized bool) {
		log.WithFields(cue.Fields{
			"user": user,
			"authorized": authorized,
		}).Info("Something else requested")
	}
```

Loggers may also be associated with object instances:

```go
	type Item struct {
		key string
		log cue.Logger
	}

	func NewItem(key string) *Item {
		return &Item{
			key: key,
			log: log.WithValue("item_key", key)
		}
	}

	func (item *Item) DoSomething() {
		// The event context will include the "item_key" set above
		item.log.Warn("Doing something important")
	}
```

## Error Logging and Recovery

Cue simplifies error reporting by logging the given error and message, and then
returning the same error value.  Hence you can return the log.Error/log.Errorf
values in-line:

```go
	filename := "somefile"
	f, err := os.Create(filename)
	if err != nil {
		return log.Errorf(err, "Failed to create %q", filename)
	}
```

Cue provides Collector implementations for popular error reporting services
such as Honeybadger, Rollbar, Sentry, and Opbeat.  If one of these collector
implementations were registered, the above code would automatically open a new
error report, complete with stack trace and context information from the logger
instance.  See the cue/hosted package for details.

Finally, cue provides convenience methods for panic and recovery. Calling Panic
or Panicf will log the provided message at the FATAL level and then panic.
Calling Recover recovers from panics and logs the recovered value and message
at the FATAL level.

```go
	func doSomething() {
		defer log.Recover("Recovered panic in doSomething")
		doSomethingThatPanics()
	}
```

If a panic is triggered via a cue logger instance's Panic or Panicf methods,
Recover recovers from the panic but suppresses the event to prevent
duplication.

## Event Collection

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

```go
	package main

	import (
		"github.com/bobziuchkovski/cue"
		"github.com/bobziuchkovski/cue/collector"
	)

	var log = cue.NewLogger("main")

	func main() {
		// Follow a 12-factor approach and log unbuffered to stdout.
		// See http://12factor.net for details.
		cue.Collect(cue.INFO, collector.Terminal{}.New())
		defer log.Recover("Recovered from panic in main")

		RunTheProgram()
	}
```

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

```go
	package main

	import (
		"github.com/bobziuchkovski/cue"
		"github.com/bobziuchkovski/cue/collector"
		"time"
	)

	var log = cue.NewLogger("main")

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
```

## Stack Frame Collection

By default, cue collects a single stack frame for any event that matches a
registered collector.  This ensures collectors may log the file name, package,
and line number for any collected event.  SetFrames may be used to alter this
frame count, or disable frame collection entirely.  See the SetFrames function
for details.

When using error reporting services, SetFrames should be used to increase the
errorFrames parameter from the default value of 1 to a value that provides
enough stack context to successfully diagnose reported errors.

```go
	package main

	import (
		"github.com/bobziuchkovski/cue"
		"github.com/bobziuchkovski/cue/collector"
		"github.com/bobziuchkovski/cue/hosted"
		"os"
		"strings"
		"time"
	)

	var log = cue.NewLogger("main")

	func main() {
		// Send WARN, ERROR, and FATAL synchronously to stdout
		cue.Collect(cue.INFO, collector.Terminal{}.New())

		// Send ERROR and FATAL asynchronously to Honeybadger, ensuring we get
		// enough context on the Honeybadger stack traces.  We use a large
		// async buffer just in case Honeybadger experiences an outage.
		cue.SetFrames(1, 32)
		cue.CollectAsync(cue.ERROR, 10000, hosted.Honeybadger{
			Key: os.Getenv("HONEYBADGER_KEY"),

		    // Optional
	        Tags: strings.Split(os.Getenv("HONEYBADGER_TAGS"), ","),
	        ExtraContext: cue.NewContext("extra").WithFields(cue.Fields{
	            "foo": "bar",
	            "frobble": 42,
	        }),
	        Environment: os.Getenv("APP_ENV"),
		}.New())

		// Close/flush buffered events on program termination.
		// Note that this won't fire if ctrl+c is used or kill <pid>.  You need
		// to install signal handlers for SIGINT/SIGTERM to handle those cases.
		defer cue.Close(5 * time.Second)

		defer log.Recover("Recovered from panic in main")
		RunTheProgram()
	}
```

Since we're using an async collector for Honeybadger, we still need cue.Close.
Otherwise a recovered panic might not post to Honeybadger prior to the main
func exiting.

There are several more items of interest in the above example:

1. We register the terminal collector before the Honeybadger collector.  This
   will surface any warnings emitted by the Honeybadger.New method.
2. If the HONEYBADGER_KEY environment variable isn't set, that collector will
   emit a warning event and return nil.  All of the built-in collectors emit
   warning log events if required parameters are missing.
3. If cue.Collect or cue.CollectAsync are called with a nil collector, they
   simply return without doing anything.  Hence the above would be safe to run
   even if HONEYBADGER_KEY is unset.

## Formatting

Collectors use Formatter functions to format message output.  The default
formatters are pretty sane, but it's easy to implement custom formats.  See the
[cue/format godocs](https://godoc.org/github.com/bobziuchkovski/cue/format)
for details.

## Colors!

People love colors.  Cue comes with a Colorize function in the format package
that wraps and colorizes an existing formatter, selecting colors by log level.
The format package also provides a HumanReadableColors pre-defined format for
conveninence.

```go
	package main

	import (
		"github.com/bobziuchkovski/cue"
		"github.com/bobziuchkovski/cue/collector"
		"github.com/bobziuchkovski/cue/format"
	)

	var log = cue.NewLogger("main")

	func main() {
		// Log to stdout...with colors!
		cue.Collect(cue.DEBUG, collector.Terminal{
			Formatter: format.HumanReadableColors,
		}.New())

		defer log.Recover("Recovered from panic in main")
		RunTheProgram()
	}
```

Please don't enable this in production.  Those escape codes buried in log
files anooy the heck out of ops/devops folk.

## Additional Examples and Docs

Please see the [godocs](https://godoc.org/github.com/bobziuchkovski/cue) for
additional examples and documentation.

## Authors

Bob Ziuchkovski (@bobziuchkovski)

## License (MIT)

Copyright (c) 2016 Bob Ziuchkovski

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.


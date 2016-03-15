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

## API Promise

Minor breaking changes may occur prior to the 1.0 release.  After the 1.0
release, the API is guaranteed to remain backwards compatible.

_Cue makes use of sync/atomic.Value and thus requires Go 1.4.x or later._

## Key Features

- Simple API.  You only need to know [cue.NewLogger](https://godoc.org/github.com/bobziuchkovski/cue#NewLogger)
  and the [cue.Logger interface](https://godoc.org/github.com/bobziuchkovski/cue#Logger) to get started.
  The rest of the API is centered around log output configuration.
- Supports both [fully synchronous](https://godoc.org/github.com/bobziuchkovski/cue#Collect) and
  [fully asynchronous](https://godoc.org/github.com/bobziuchkovski/cue#CollectAsync) (guaranteed non-blocking) log collection.
- Supports a variety of log outputs:
  * [File](https://godoc.org/github.com/bobziuchkovski/cue/collector#File)
  * [Syslog](https://godoc.org/github.com/bobziuchkovski/cue/collector#Syslog)
  * [Structured Syslog](https://godoc.org/github.com/bobziuchkovski/cue/collector#StructuredSyslog)
  * [Stdout/Stderr](https://godoc.org/github.com/bobziuchkovski/cue/collector#Terminal)
  * [Socket](https://godoc.org/github.com/bobziuchkovski/cue/collector#Socket)
  * [Honeybadger](https://godoc.org/github.com/bobziuchkovski/cue/hosted#Honeybadger)
  * [Loggly](https://godoc.org/github.com/bobziuchkovski/cue/hosted#Loggly)
  * [Opbeat](https://godoc.org/github.com/bobziuchkovski/cue/hosted#Opbeat)
  * [Rollbar](https://godoc.org/github.com/bobziuchkovski/cue/hosted#Rollbar)
  * [Sentry](https://godoc.org/github.com/bobziuchkovski/cue/hosted#Sentry)
- Very flexible [formatting](https://godoc.org/github.com/bobziuchkovski/cue/format)
- Designed to stay out of your way.  Log collection is explicitly opt-in, meaning cue is safe to use within
  libraries.  If the end user doesn't configure log collection, logging calls are silently dropped.
- Designed with performance in mind.  Cue uses atomic operations to avoid lock contention and reuses
  formatting buffers to avoid excessive memory allocation and garbage collection.  Logging calls that don't
  match a registered Collector threshold (e.g. DEBUG in production) are quickly pruned via atomic operations.
  On a 2015 MacBook pro, these calls take ~16ns/call, meaning tens of millions of no-op calls may be serviced
  per second.
- Gracefully [handles and recovers from failures](https://godoc.org/github.com/bobziuchkovski/cue/collector#hdr-Collector_Failures_and_Degradation).
- Has thorough test coverage:
  * [cue](https://godoc.org/github.com/bobziuchkovski/cue): [![Coverage](https://gocover.io/_badge/github.com/bobziuchkovski/cue?1)](https://gocover.io/github.com/bobziuchkovski/cue)
  * [cue/collector](https://godoc.org/github.com/bobziuchkovski/cue/collector): [![Coverage](https://gocover.io/_badge/github.com/bobziuchkovski/cue/collector?1)](https://gocover.io/github.com/bobziuchkovski/cue/collector)
  * [cue/format](https://godoc.org/github.com/bobziuchkovski/cue/format): [![Coverage](https://gocover.io/_badge/github.com/bobziuchkovski/cue/format?1)](https://gocover.io/github.com/bobziuchkovski/cue/format)
  * [cue/hosted](https://godoc.org/github.com/bobziuchkovski/cue/hosted): [![Coverage](https://gocover.io/_badge/github.com/bobziuchkovski/cue/hosted?1)](https://gocover.io/github.com/bobziuchkovski/cue/hosted)

## Basic Use

Please see the [godocs](https://godoc.org/github.com/bobziuchkovski/cue) for
additional information.

This example logs to both the terminal (stdout) and to file. If the program
receives SIGHUP, the file will be reopened.  This is useful for log rotation.
Additional context is added via the .WithValue and .WithFields Logger methods.

The formatting may be changed by passing a different formatter to either collector.
See the [cue/format godocs](https://godoc.org/github.com/bobziuchkovski/cue/format)
for details.  The context data may also be formatted as JSON for machine parsing
if desired.  See cue/format.JSONMessage and cue/format.JSONContext.

```go
package main

import (
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/collector"
	"os"
	"syscall"
)

var log = cue.NewLogger("main")

func main() {
	cue.Collect(cue.INFO, collector.Terminal{}.New())
	cue.Collect(cue.INFO, collector.File{
		Path:         "app.log",
		ReopenSignal: syscall.SIGHUP,
	}.New())

	log := cue.NewLogger("example")
	log.Debug("Debug message -- a quick no-op since our collector is registered at INFO level")
	log.Info("Info message")
	log.Warn("Warn message")

	// Add additional context
	log.WithValue("items", 2).Infof("This is an %s", "example")
	log.WithFields(cue.Fields{
		"user":          "bob",
		"authenticated": true,
	}).Warn("Doing something important")

	host, err := os.Hostname()
	if err != nil {
		log.Error(err, "Failed to retrieve hostname")
	} else {
		log.Infof("My hostname is %s", host)
	}

	// The output looks something like:
	// Mar 13 12:40:10 INFO example_basic_test.go:25 Info message
	// Mar 13 12:40:10 WARN example_basic_test.go:26 Warn message
	// Mar 13 12:40:10 INFO example_basic_test.go:29 This is an example items=2
	// Mar 13 12:40:10 WARN example_basic_test.go:33 Doing something important user=bob authenticated=true
	// Mar 13 12:40:10 INFO example_basic_test.go:39 My hostname is pegasus.bobbyz.org
}
```

## Error Reporting

Please see the [godocs](https://godoc.org/github.com/bobziuchkovski/cue) for
additional information.

This example uses cue/hosted.Honeybadger to report error events to Honeybadger.

```go
package main

import (
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/hosted"
	"os"
	"time"
)

var log = cue.NewLogger("main")

func main() {
	// Here we're assuming the Honeybadger API key is stored via environment
	// variable, as well as an APP_ENV variable specifying "test", "production", etc.
	cue.CollectAsync(cue.ERROR, 10000, hosted.Honeybadger{
		Key:         os.Getenv("HONEYBADGER_KEY"),
		Environment: os.Getenv("APP_ENV"),
	}.New())

	// We want to collect more stack frames for error and panic events so that
	// our Honeybadger incidents show us enough stack trace to troubleshoot.
	cue.SetFrames(1, 32)

	// We use Close to flush the asynchronous buffer.  This way we won't
	// lose error reports if one is in the process of sending when the program
	// is terminating.
	defer cue.Close(5 * time.Second)

	// If something panics, it will automatically open a Honeybadger event
	// when recovered by this line
	defer log.Recover("Recovered panic")

	// Force a panic
	PanickingFunc()
}

func PanickingFunc() {
	panic("This will be reported to Honeybadger")
}
```

## Features

Please see the [godocs](https://godoc.org/github.com/bobziuchkovski/cue) for
additional information.

This example shows quite a few of the cue features: logging to a file that
reopens on SIGHUP (for log rotation), logging colored output to stdout,
logging to syslog with JSON context formatting, and reporting errors to
Honeybadger.

```go
package main

import (
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/collector"
	"github.com/bobziuchkovski/cue/format"
	"github.com/bobziuchkovski/cue/hosted"
	"os"
	"syscall"
	"time"
)

var log = cue.NewLogger("main")

func main() {
	// defer cue.Close before log.Recover so that Close flushes any events
	// triggers by panic recovery
	defer cue.Close(5 * time.Second)
	defer log.Recover("Recovered panic in main")
	ConfigureLogging()
	RunTheProgram()
}

func ConfigureLogging() {
	// Collect logs to stdout in color!  :)
	cue.Collect(cue.DEBUG, collector.Terminal{
		Formatter: format.HumanReadableColors,
	}.New())

	// Collect to app.log and reopen the handle if we receive SIGHUP
	cue.Collect(cue.INFO, collector.File{
		Path:         "app.log",
		ReopenSignal: syscall.SIGHUP,
	}.New())

	// Collect to syslog, formatting the context data as JSON for indexing.
	cue.Collect(cue.WARN, collector.Syslog{
		App:       "app",
		Facility:  collector.LOCAL7,
		Formatter: format.JSONMessage,
	}.New())

	// Report errors asynchronously to Honeybadger.  If HONEYBADGER_KEY is
	// unset, Honeybadger.New will return nil and cue.CollectAsync will
	// ignore it.  This works great for development.
	cue.CollectAsync(cue.ERROR, 10000, hosted.Honeybadger{
		Key:         os.Getenv("HONEYBADGER_KEY"),
		Environment: os.Getenv("APP_ENV"),
	}.New())
	cue.SetFrames(1, 32)
}

func RunTheProgram() {
	log.Info("Running the program!")
	log.WithFields(cue.Fields{
		"sad":    true,
		"length": 0,
	}).Panic("No program", "Whoops, there's no program to run!")
}
```

## Motivation

There are quite a few Go logging libraries that already exist.  Why create a new one?

I didn't start with the intention of creating a new library, but I struggled to find an existing logging library that
met my primary objectives:

1. **Built-in support for error reporting services**.  Error reporting services are incredibly valuable for production
  applications.  They provide immediate notification and visibility into bugs that your end users are encountering.
  However, there's nothing particularly *special* about how error reporting services work.  They are merely a form of error logging.
  I feel it's important to be able to log/collect application errors without needing to explicitly specify where to send them.
  Calling log.Error/log.Fatal/log.Recover should be enough to trigger the error report.  Where that error is sent should be
  a configuration detail and shouldn't require additional service-specific API calls.
2. **Contextual logging support**.  It's often valuable to know under what conditions application events trigger: the account that
  accessed a page, flags associated with the account, country of origin, etc.  Cue simplifies collection of these details via
  the Logger.WithFields and Logger.WithValues methods, and provides flexible formatting options for rendering these values,
  including machine-parseable formats for search and indexing.  Furthermore, I have every intention of providing a bridge API to
  integrate contextual logging with the upstream "context" package depending on what happens with [this proposal](https://github.com/golang/go/issues/14660).
3. **Asynchronous logging with guaranteed non-blocking behavior**.  When logging to third-party services, you don't want a
  service disruption to halt your entire application.  Unfortunately, many existing logging libraries obtain a global lock
  for logging calls and log synchronously.  This is low risk when logging to stdout, but high risk when logging to external
  services.  If those external services go down, your entire application may halt trying to acquire a logging mutex that's
  blocked trying to write to the downed service.
4. **Low overhead**.  Developers should feel comfortable peppering their code with logging calls wherever there's value
  in collecting the data.  They shouldn't need to worry about the overhead of those logging calls or bottlenecking on
  logging.  To that point, cue is carefully designed to avoid lock contention and to avoid excessive object allocation.
  Cue uses atomic operations to quickly prune logging calls that don't match registered collector thresholds
  (e.g. DEBUG logs in production).  Cue also reuses formatting buffers via a sync.Pool to reduce object allocations and
  garbage collection overhead.  Logging calls aren't free, but cue makes them as cheap as possible.
5. **Has thorough test coverage**.  If you're peppering your code with calls to a logging API, you should feel comfortable
  that the API is well-tested and stable.  Many existing logging libraries have weak test coverage and lack API promises.
6. **Friendly to library authors**.  Many of the existing logging libraries write log output to stdout/stderr
  or to file by default.  This makes life difficult on library authors because you're left either interfering
  with your end user's logging configuration (explicitly changing logging library defaults), or omitting logging
  altogether.  Cue addresses this by requiring explicit opt-in *by the end user* in order to collect logging output.
  Uncollected logging calls are no-ops.

## Documentation

Please see the [godocs](https://godoc.org/github.com/bobziuchkovski/cue) for
additional details.

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

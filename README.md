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

Please see the [godocs](https://godoc.org/github.com/bobziuchkovski/cue) for
additional information.

This example shows how to register the terminal collector (stdout) and log a few
messages at various levels.

```go
package main

import (
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/collector"
	"os"
)

var log = cue.NewLogger("main")

func main() {
	cue.Collect(cue.INFO, collector.Terminal{}.New())

	log.Debug("Debug message -- a quick no-op since our collector is registered at INFO level")
	log.Info("Info message")
	log.Warn("Warn message")

	host, err := os.Hostname()
	if err != nil {
		log.Error(err, "Failed to retrieve hostname")
	} else {
		log.Infof("My hostname is %s", host)
	}

	// The output looks something like:
	// Mar 13 12:40:10 INFO example_basic_test.go:20 Info message
	// Mar 13 12:40:10 WARN example_basic_test.go:21 Warn message
	// Mar 13 12:40:10 INFO example_basic_test.go:27 My hostname is pegasus.bobbyz.org

	// The formatting could be changed by passing a different formatter to collector.Terminal.
	// see the cue/format docs for details
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
logging to syslog, and reporting errors to Honeybadger.

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

	// Collect to syslog
	cue.Collect(cue.WARN, collector.Syslog{
		App:      "app",
		Facility: collector.LOCAL7,
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
	panic("Whoops, there's no program to run!")
}
```


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

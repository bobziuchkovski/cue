// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cue_test

import (
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/collector"
	"github.com/bobziuchkovski/cue/format"
	"github.com/bobziuchkovski/cue/hosted"
	"os"
	"syscall"
	"time"
)

var log = cue.NewLogger("example")

// This example shows quite a few of the cue features: logging to a file that
// reopens on SIGHUP (for log rotation), logging colored output to stdout,
// logging to syslog, and reporting errors to Honeybadger.
func Example_features() {
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

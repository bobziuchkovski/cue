// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cue_test

import (
	"os"
	"syscall"

	"github.com/remerge/cue"
	"github.com/remerge/cue/collector"
)

// This example logs to both the terminal (stdout) and to file.
// If the program receives SIGHUP, the file will be reopened (for log rotation).
// Additional context is added via the .WithValue and .WithFields Logger methods.
//
// The formatting may be changed by passing a different formatter to either collector.
// See the cue/format godocs for details.  The context data may also be formatted as
// JSON for machine parsing if desired.  See cue/format.JSONMessage and cue/format.JSONContext.
func Example_basic() {
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

	// The formatting could be changed by passing a different formatter to collector.Terminal.
	// see the cue/format docs for details
}

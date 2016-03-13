// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cue_test

import (
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/collector"
	"os"
)

// This example shows how to register the terminal collector and log a few
// messages at various levels.
func Example_basic() {
	cue.Collect(cue.INFO, collector.Terminal{}.New())

	log := cue.NewLogger("example")
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

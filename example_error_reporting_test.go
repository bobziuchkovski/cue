// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cue_test

import (
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/hosted"
	"os"
	"time"
)

// This example shows how to use error reporting services.
func Example_errorReporting() {
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
	log := cue.NewLogger("example")
	defer log.Recover("Recovered panic")

	// Force a panic
	PanickingFunc()
}

func PanickingFunc() {
	panic("This will be reported to Honeybadger")
}

// Copyright 2016 Bob Ziuchkovski. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package hosted_test

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

	"github.com/remerge/cue"
	"github.com/remerge/cue/collector"
	"github.com/remerge/cue/hosted"
)

// This example demonstrates how to use Loggly with TLS transport encryption
// enabled.
func ExampleLoggly_transportEncryption() {
	// Load Loggly's CA cert.  Please see the Loggly docs for details on
	// retrieving their cert files.
	pem, err := ioutil.ReadFile("sf_bundle.crt")
	if err != nil {
		panic(err)
	}
	cacert := x509.NewCertPool()
	if !cacert.AppendCertsFromPEM(pem) {
		panic("failed to load loggly CA cert")
	}

	cue.Collect(cue.INFO, hosted.Loggly{
		Token:    os.Getenv("LOGGLY_TOKEN"),
		App:      "example_loggly_tls",
		Facility: collector.LOCAL0,
		Network:  "tcp",
		Address:  "logs-01.loggly.com:6514", // Loggly uses port 6514 for TLS
		TLS:      &tls.Config{RootCAs: cacert},
	}.New())

	log := cue.NewLogger("example")
	log.Info("This event is sent over TLS")
}

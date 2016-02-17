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
Package hosted implements event collection for hosted third-party services.
Collectors are provided for Honeybadger, Loggly, Opbeat, Rollbar, and Sentry.
Additional collectors will be added upon request.

Inclusion Criteria

The following criteria are used to evaluate third-party services:

1. Does the service provide a perpetual free tier?  This is a must-have
requirement.

2. Does the service offer transport security?  This is a must-have
requirement.

3. Is the service a good fit for collecting cue events?  Logging and error
reporting services are a great fit. E-mail and messaging services, on the
other hand, are intentionally omitted.

4. Does the service provide a sane API/integration mechanism?  Firing JSON
HTTP posts is simple, whereas implementing proprietary transport mechanisms
is a pain.

If a third-party service meets the above criteria and isn't supported, feel
free to open a feature request.

Frame Collection

By default, cue collects a single stack frame for all logged events.
Increasing the number of frames collected for ERROR and FATAL events is a
good idea when using error reporting services.  See the cue.SetFrames docs for
details.

Nil Instances

Collector implementations emit a WARN log event and return a nil collector
instance if required parameters are missing.  The cue.Collect and
cue.CollectAsync functions treat nil collectors as a no-op, so this is
perfectly safe.
*/
package hosted

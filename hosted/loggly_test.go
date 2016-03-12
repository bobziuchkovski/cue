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

package hosted

import (
	"fmt"
	"github.com/bobziuchkovski/cue/collector"
	"github.com/bobziuchkovski/cue/internal/cuetest"
	"reflect"
	"regexp"
	"testing"
)

func TestLogglyNilCollector(t *testing.T) {
	c := Loggly{App: "test"}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the token is missing, but got %s instead", c)
	}

	c = Loggly{Token: "test"}.New()
	if c != nil {
		t.Errorf("Expected a nil collector when the app is missing, but got %s instead", c)
	}
}

func TestLogglyDefaultHostNet(t *testing.T) {
	c := Loggly{Token: "test", App: "test"}.New()
	if c == nil {
		t.Error("Expected to get a non-nil collector with Token and App specified, but got a nil collector instead")
	}
}

func TestLoggly(t *testing.T) {
	recorder := cuetest.NewTCPRecorder()
	recorder.Start()
	defer recorder.Close()

	c := getLogglyCollector("tcp", recorder.Address())

	err := c.Collect(cuetest.DebugEvent)
	if err != nil {
		t.Errorf("Encountered unexpected error: %s", err)
	}
	cuetest.CloseCollector(c)

	pattern := `<167>1 2006-01-02T15:04:00.000000(Z|[-+]\d{2}:\d{2}) \S+ testapp testapp\[\d+\] - \[test@41058 tag="tag1" tag="tag2"\] debug event {"k1":"some value","k2":2,"k3":3.5,"k4":true}\n`
	re := regexp.MustCompile(pattern)

	if !re.Match(recorder.Contents()) {
		t.Errorf("Expected content %q to match pattern %q but it didn't", recorder.Contents(), pattern)
	}
}

func TestLogglyString(t *testing.T) {
	_ = fmt.Sprint(getLogglyCollector("tcp", "localhost:12345"))
}

func getLogglyCollector(net, addr string) *logglyCollector {
	c := Loggly{
		Token:    "test",
		App:      "testapp",
		Tags:     []string{"tag1", "tag2"},
		Facility: collector.LOCAL4,
		Network:  net,
		Address:  addr,
	}.New()
	lc, ok := c.(*logglyCollector)
	if !ok {
		panic(fmt.Sprintf("Expected to see a *logglyCollector but got %s instead", reflect.TypeOf(c)))
	}
	return lc
}

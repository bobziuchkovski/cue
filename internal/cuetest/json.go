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

package cuetest

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// ParseRequestJSON parses the request body as json, returning a
// nested map[string]interface{} of the decoded contents.
// If the content isn't well-formed, ParseRequestJSON panics.
func ParseRequestJSON(req *http.Request) map[string]interface{} {
	j := make(map[string]interface{})
	d := json.NewDecoder(req.Body)
	d.UseNumber()
	err := d.Decode(&j)
	if err != nil {
		panic(err)
	}
	return j
}

// ParseStringJSON parses the input jstr as json, returning a
// nested map[string]interface{} of the decoded contents.
// If the content isn't well-formed ParseStringJSON panics.
func ParseStringJSON(jstr string) map[string]interface{} {
	j := make(map[string]interface{})
	d := json.NewDecoder(strings.NewReader(jstr))
	d.UseNumber()
	err := d.Decode(&j)
	if err != nil {
		panic(err)
	}
	return j
}

// NestedFetch treats j as a nested map[string]interface{} and attempts to
// retrieve the value specified by path.  It returns "!(MISSING)" if the
// value is missing, or "!(NOTAKEY)" if part of the path exists but terminates
// early at a value that isn't a key.
func NestedFetch(j map[string]interface{}, path ...string) interface{} {
	for i, part := range path {
		v, present := j[part]
		if !present {
			return "!(MISSING)"
		}
		if i >= len(path)-1 {
			return v
		}
		sub, ok := v.(map[string]interface{})
		if !ok {
			return "!(NOTAKEY)"
		}
		j = sub
	}
	return "!(MISSING)"
}

// NestedDelete treats j as a nested map[string]interface{} and attempts to
// delete the value specified by the path.  If does nothing if the path doesn't
// correspond to a valid key.
func NestedDelete(j map[string]interface{}, path ...string) {
	for i, part := range path {
		v, present := j[part]
		if !present {
			return
		}
		if i >= len(path)-1 {
			delete(j, part)
			return
		}
		sub, ok := v.(map[string]interface{})
		if !ok {
			return
		}
		j = sub
	}
}

// NestedCompare treats input and expected as nested map[string]interface{} and
// performs a deep comparison between them.  If the maps aren't equal, it
// calls t.Errorf with a pretty-printed comparison.
func NestedCompare(t *testing.T, input map[string]interface{}, expected map[string]interface{}) {
	if !reflect.DeepEqual(input, expected) {
		prettyInput := prettyFormat(input)
		prettyExpected := prettyFormat(expected)
		t.Errorf(`
Request JSON doesn't match expectations.

Expected
========
%s

Received
========
%s
`, prettyExpected, prettyInput)
	}
}

func prettyFormat(j map[string]interface{}) string {
	bytes, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

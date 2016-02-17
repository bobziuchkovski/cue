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

package cue

import (
	"runtime"
	"strings"
)

// Frame fields use UnknownPackage, UnknownFunction, and UnknownFile when the
// package, function, or file cannot be determined for a stack frame.
const (
	UnknownPackage  = "<unknown package>"
	UnknownFunction = "<unknown function>"
	UnknownFile     = "<unknown file>"
)

var nilFrame = &Frame{
	Package:  UnknownPackage,
	Function: UnknownFunction,
	File:     UnknownFile,
	Line:     0,
}

// Frame represents a single stack frame.
type Frame struct {
	Package  string // Package name or cue.UnknownPackage ("<unknown package>") if unknown
	Function string // Function name or cue.UnknownFunction ("<unknown function>") if unknown
	File     string // Full file path or cue.UnknownFile ("<unknown file>") if unknown
	Line     int    // Line Number or 0 if unknown
}

func frameForPC(pc uintptr) *Frame {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return nilFrame
	}

	file, line := fn.FileLine(pc)
	function := fn.Name()
	return &Frame{
		Package:  packageForFunc(function),
		Function: function,
		File:     file,
		Line:     line,
	}
}

func packageForFunc(fn string) string {
	pkg := fn
	slashidx := strings.LastIndex(pkg, "/")
	if slashidx == -1 {
		slashidx = 0
	}
	dotidx := strings.Index(pkg[slashidx:], ".")
	if dotidx == -1 {
		dotidx = len(pkg)
	}
	return pkg[:slashidx+dotidx]
}

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

package format

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bobziuchkovski/cue"
)

// Color codes for use with Colorize.
const (
	red    = 31
	green  = 32
	yellow = 33
	blue   = 34
)

// Pre-defined Formatters.  HumanReadable is a nice default when machine
// parsing isn't required.
var (
	// Message[: Error] key1=val1 key2=val2...
	HumanMessage = Escape(Trim(Join(" ", MessageWithError, HumanContext)))

	// Jan _2 15:04:05 INFO [Shortfile:Line] Message[: Error] key1=val1 key2=val2...
	HumanReadable       = Join(" ", Time(time.Stamp), Level, SourceWithLine, HumanMessage)
	HumanReadableColors = Colorize(HumanReadable)

	// Message[: Error] {"key1":"val1","key2":"val2"}
	JSONMessage = Join(" ", Escape(Trim(MessageWithError)), JSONContext)
)

// Formatter is the interface used to format Collector output.
type Formatter func(buffer Buffer, event *cue.Event)

// RenderBytes renders the given event using formatter.
func RenderBytes(formatter Formatter, event *cue.Event) []byte {
	tmp := GetBuffer()
	defer ReleaseBuffer(tmp)

	formatter(tmp, event)
	result := make([]byte, tmp.Len())
	copy(result, tmp.Bytes())
	return result
}

// RenderString renders the given event using formatter.
func RenderString(formatter Formatter, event *cue.Event) string {
	tmp := GetBuffer()
	defer ReleaseBuffer(tmp)

	formatter(tmp, event)
	return string(tmp.Bytes())
}

// Join returns a new Formatter that appends sep between the contents of
// underlying formatters.  Sep is only appended between formatters that write
// one or more bytes to their buffers.
func Join(sep string, formatters ...Formatter) Formatter {
	return func(buffer Buffer, event *cue.Event) {
		tmp := GetBuffer()
		defer ReleaseBuffer(tmp)

		needSep := false
		for _, formatter := range formatters {
			formatter(tmp, event)
			if tmp.Len() == 0 {
				continue
			}

			if needSep {
				buffer.AppendString(sep)
			}
			buffer.Append(tmp.Bytes())
			tmp.Reset()
			needSep = true
		}
	}
}

// Formatf provides printf-like formatting of source formatters. The "%v"
// placeholder is used to specify formatter placeholders.  In the rare event
// a literal "%v" is required, "%%v" renders the literal.  No alignment,
// padding, or other printf constructs are currently supported, though code
// contributions are certainly welcome.
func Formatf(format string, formatters ...Formatter) Formatter {
	formatterIdx := 0
	segments := splitFormat(format)
	chain := make([]Formatter, len(segments))
	for i, seg := range segments {
		switch {
		case seg == "%v" && formatterIdx < len(formatters):
			chain[i] = formatters[formatterIdx]
			formatterIdx++
		case seg == "%v":
			chain[i] = Literal("%!v(MISSING)")
		default:
			chain[i] = Literal(seg)
		}
	}

	return func(buffer Buffer, event *cue.Event) {
		for _, formatter := range chain {
			formatter(buffer, event)
		}
	}
}

func splitFormat(format string) []string {
	var (
		segments []string
		segstart int
		lastrune rune
	)

	runes := []rune(format)
	for i, r := range runes {
		switch {
		case lastrune == '%' && r == '%':
			segend := i - 1
			if segstart != segend {
				segments = append(segments, string(runes[segstart:segend]))
			}
			segments = append(segments, "%")
			segstart = i + 1
			lastrune = 0
		case lastrune == '%' && r == 'v':
			segend := i - 1
			if segstart != segend {
				segments = append(segments, string(runes[segstart:segend]))
			}
			segments = append(segments, "%v")
			segstart = i + 1
			lastrune = r
		default:
			lastrune = r
		}
	}

	if segstart < len(runes) {
		segments = append(segments, string(runes[segstart:]))
	}
	return segments
}

// Colorize returns a new formatter that wraps the underlying formatter output
// in color escape codes by level: DEBUG output is blue, INFO output is green,
// WARN output is yellow, and ERROR/FATAL output is red.  No additional color
// support is provided, nor will any be added.
func Colorize(formatter Formatter) Formatter {
	return func(buffer Buffer, event *cue.Event) {
		buffer.AppendString(fmt.Sprintf("\x1b[%dm", colorFor(event.Level)))
		formatter(buffer, event)
		buffer.AppendString("\x1b[0m")
	}
}

func colorFor(lvl cue.Level) int {
	switch lvl {
	case cue.DEBUG:
		return blue
	case cue.INFO:
		return green
	case cue.WARN:
		return yellow
	case cue.ERROR, cue.FATAL:
		return red
	default:
		panic("cue/format: BUG unknown level")
	}
}

// Trim returns a formatter that trims leading and trailing whitespace from
// the input formatter.
func Trim(formatter Formatter) Formatter {
	return func(buffer Buffer, event *cue.Event) {
		tmp := GetBuffer()
		defer ReleaseBuffer(tmp)

		formatter(tmp, event)
		buffer.AppendString(strings.TrimSpace(string(tmp.Bytes())))
	}
}

// Escape returns a formatter that escapes all control characters and all
// whitespace characters other than ' ' (ASCII space) from the input formatter.
func Escape(formatter Formatter) Formatter {
	return func(buffer Buffer, event *cue.Event) {
		tmp := GetBuffer()
		defer ReleaseBuffer(tmp)

		formatter(tmp, event)
		runes := []rune(string(tmp.Bytes()))
		for _, r := range runes {
			switch {
			case r == ' ':
				buffer.AppendRune(r)
			case unicode.IsControl(r), unicode.IsSpace(r):
				quoted := strconv.QuoteRune(r)
				buffer.AppendString(quoted[1 : len(quoted)-1])
			default:
				buffer.AppendRune(r)
			}
		}
	}
}

// Truncate returns a new formatter that truncates the input formatter after
// length bytes are written.
func Truncate(formatter Formatter, length int) Formatter {
	return func(buffer Buffer, event *cue.Event) {
		tmp := GetBuffer()
		defer ReleaseBuffer(tmp)

		formatter(tmp, event)
		bytes := tmp.Bytes()
		if len(bytes) > length {
			bytes = bytes[:length]
		}
		buffer.Append(bytes)
	}
}

// Literal returns a formatter that always writes s to its buffer.
func Literal(s string) Formatter {
	return func(buffer Buffer, event *cue.Event) {
		buffer.AppendString(s)
	}
}

// Time returns a formatter that writes the event's timestamp to the buffer
// using the formatting rules from the time package.
func Time(timeFormat string) Formatter {
	return func(buffer Buffer, event *cue.Event) {
		buffer.AppendString(event.Time.Format(timeFormat))
	}
}

// Hostname writes the host's short name to the buffer, domain excluded.
// If the hostname cannot be determined, "unknown" is written instead.
func Hostname(buffer Buffer, event *cue.Event) {
	name, err := os.Hostname()
	if err != nil {
		name = "unknown"
	}
	buffer.AppendString(name)
}

// FQDN writes the host's fully-qualified domain name (FQDN) to the buffer.
// If the FQDN cannot be determined, "unknown" is written instead.
func FQDN(buffer Buffer, event *cue.Event) {
	out, err := exec.Command("/bin/hostname", "-f").Output()
	if err == nil {
		buffer.Append(bytes.TrimSpace(out))
	} else {
		buffer.AppendString("unknown")
	}
}

// Level writes event.Level.String() to the buffer.  Hence, it writes "INFO"
// for INFO level messages, "DEBUG" for DEBUG level messages, and so on.
func Level(buffer Buffer, event *cue.Event) {
	buffer.AppendString(event.Level.String())
}

// Package writes the package name that generated the event.  If this cannot
// be determined or frame collection is disabled, it writes cue.UnknownPackage
// ("<unknown package>") instead.
func Package(buffer Buffer, event *cue.Event) {
	if len(event.Frames) == 0 {
		buffer.AppendString(cue.UnknownPackage)
		return
	}
	buffer.AppendString(event.Frames[0].Package)
}

// Function writes the function name that generated the event.  If this cannot
// be determined or frame collection is disabled, it writes cue.UnknownFunction
// ("<unknown function>") instead.
func Function(buffer Buffer, event *cue.Event) {
	if len(event.Frames) == 0 {
		buffer.AppendString(cue.UnknownFunction)
		return
	}
	buffer.AppendString(event.Frames[0].Function)
}

// File writes the source file name that generated the event, path included.
// If this cannot be determined or frame collection is disabled, it writes
// cue.UnknownFile ("<unknown file>") instead.
func File(buffer Buffer, event *cue.Event) {
	if len(event.Frames) == 0 {
		buffer.AppendString(cue.UnknownFile)
		return
	}
	buffer.AppendString(event.Frames[0].File)
}

// ShortFile writes the source file name that generated the event, path
// omitted.  If this cannot be determined or frame collection is disabled,
// it writes cue.UnknownFile ("<unknown file>") instead.
func ShortFile(buffer Buffer, event *cue.Event) {
	if len(event.Frames) == 0 {
		buffer.AppendString(cue.UnknownFile)
		return
	}
	short := event.Frames[0].File
	idx := strings.LastIndex(short, "/")
	if idx != -1 {
		short = short[idx+1:]
	}
	buffer.AppendString(short)
}

// Line writes the source line number that generated the event. If this
// cannot be determined or frame collection is disabled, it writes "0"
// instead.
func Line(buffer Buffer, event *cue.Event) {
	if len(event.Frames) == 0 {
		buffer.AppendString("0")
		return
	}
	buffer.AppendString(fmt.Sprintf("%d", event.Frames[0].Line))
}

// Message writes event.Message to the buffer.
func Message(buffer Buffer, event *cue.Event) {
	buffer.AppendString(event.Message)
}

// Error writes event.Error.Error() to the buffer.  If event.Error is nil,
// nothing is written.
func Error(buffer Buffer, event *cue.Event) {
	if event.Error == nil {
		return
	}
	buffer.AppendString(event.Error.Error())
}

// ErrorType writes the dereferenced type name for the event's Error field.
// If event.Error is nil, nothing is written.
func ErrorType(buffer Buffer, event *cue.Event) {
	if event.Error == nil {
		return
	}
	rtype := reflect.TypeOf(event.Error)
	for rtype.Kind() == reflect.Ptr {
		rtype = rtype.Elem()
	}
	buffer.AppendString(rtype.String())
}

// MessageWithError writes event.Message to the buffer, followed by ": " and
// event.Error.Error().  The latter portions are omitted if event.Error is nil.
func MessageWithError(buffer Buffer, event *cue.Event) {
	buffer.AppendString(event.Message)
	if event.Error != nil && event.Error.Error() != event.Message {
		buffer.AppendString(": ")
		buffer.AppendString(event.Error.Error())
	}
}

// SourceWithLine writes ShortFile, followed by ":" and Line.  If these cannot
// be determined or frame collection is disabled, nothing is written.
func SourceWithLine(buffer Buffer, event *cue.Event) {
	short := RenderString(ShortFile, event)
	if short == cue.UnknownFile {
		return
	}
	buffer.AppendString(short)
	buffer.AppendRune(':')
	buffer.AppendString(RenderString(Line, event))
}

// ContextName writes event.Context.Name() to the buffer.  This is the name
// provided to cue.NewLogger().
func ContextName(buffer Buffer, event *cue.Event) {
	buffer.AppendString(event.Context.Name())
}

// HumanContext writes the event.Context key/value pairs in key=value format.
// This is similar to the format for structured logging prescribed by RFC5424,
// but suppresses quotes on values that don't contain spaces, quotes, or
// control characters.  Other values are quoted using strconv.Quote.
func HumanContext(buffer Buffer, event *cue.Event) {
	fields := event.Context.Fields()

	// Sort field keys for predictable output ordering
	var sortedKeys []string
	for k := range fields {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for i, k := range sortedKeys {
		writeHumanValue(buffer, k)
		buffer.AppendRune('=')
		writeHumanValue(buffer, fields[k])
		if i < len(sortedKeys)-1 {
			buffer.AppendRune(' ')
		}
	}
}

func writeHumanValue(buffer Buffer, v interface{}) {
	s := fmt.Sprint(v)
	if len(s) == 0 {
		buffer.AppendString(`""`)
		return
	}

	special := func(r rune) bool {
		switch {
		case r == '"', r == '\'', r == '\\', r == 0:
			return true
		case unicode.IsLetter(r), unicode.IsNumber(r), unicode.IsPunct(r), unicode.IsSymbol(r):
			return false
		default:
			return true
		}
	}
	if strings.IndexFunc(s, special) >= 0 {
		buffer.AppendString(strconv.Quote(s))
		return
	}
	buffer.AppendString(s)
}

// JSONContext marshals the event.Context fields into JSON and writes the
// result.
func JSONContext(buffer Buffer, event *cue.Event) {
	fields := event.Context.Fields()
	marshaled, _ := json.Marshal(fields)
	buffer.Append(marshaled)
}

// StructuredContext marshals the event.Context fields into structured
// key=value pairs as prescribed by RFC 5424, "The Syslog Protocol".
func StructuredContext(buffer Buffer, event *cue.Event) {
	tmp := GetBuffer()
	defer ReleaseBuffer(tmp)

	needSep := false
	event.Context.Each(func(name string, value interface{}) {
		if !validStructuredKey(name) {
			return
		}

		writeStructuredPair(tmp, name, value)
		if needSep {
			buffer.AppendRune(' ')
		}
		buffer.Append(tmp.Bytes())
		tmp.Reset()
		needSep = true
	})
}

// These restrictions are imposed by RFC 5424.
func validStructuredKey(name string) bool {
	if len(name) > 32 {
		return false
	}
	for _, r := range []rune(name) {
		switch {
		case r <= 32:
			return false
		case r >= 127:
			return false
		case r == '=', r == ']', r == '"':
			return false
		}
	}
	return true
}

func writeStructuredPair(buffer Buffer, name string, value interface{}) {
	buffer.AppendString(name)
	buffer.AppendRune('=')
	buffer.AppendRune('"')
	writeStructuredValue(buffer, value)
	buffer.AppendRune('"')
}

// See Section 6.3.3 of RFC 5424 for details on the character escapes
func writeStructuredValue(buffer Buffer, v interface{}) {
	s, ok := v.(string)
	if !ok {
		s = fmt.Sprint(v)
	}

	for _, r := range []rune(s) {
		switch r {
		case '"':
			buffer.AppendRune('\\')
			buffer.AppendRune('"')
		case '\\':
			buffer.AppendRune('\\')
			buffer.AppendRune('\\')
		case ']':
			buffer.AppendRune('\\')
			buffer.AppendRune(']')
		default:
			buffer.AppendRune(r)
		}
	}
}

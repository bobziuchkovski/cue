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
	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/internal/cuetest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRenderBytes(t *testing.T) {
	b := RenderBytes(Literal("test"), cuetest.DebugEvent)
	checkRendered(t, "test", string(b))
}

func TestRenderString(t *testing.T) {
	s := RenderString(Literal("test"), cuetest.DebugEvent)
	checkRendered(t, "test", s)
}

func TestHumanMessage(t *testing.T) {
	expected := `debug event k1="some value" k2=2 k3=3.5 k4=true`
	checkRendered(t, expected, RenderString(HumanMessage, cuetest.DebugEvent))

	expected = `error event: error message k1="some value" k2=2 k3=3.5 k4=true`
	checkRendered(t, expected, RenderString(HumanMessage, cuetest.ErrorEvent))
}

func TestHumanReadable(t *testing.T) {
	expected := `Jan  2 15:04:00 DEBUG debug event k1="some value" k2=2 k3=3.5 k4=true`
	checkRendered(t, expected, RenderString(HumanReadable, cuetest.DebugEventNoFrames))

	expected = `Jan  2 15:04:00 DEBUG file3.go:3 debug event k1="some value" k2=2 k3=3.5 k4=true`
	checkRendered(t, expected, RenderString(HumanReadable, cuetest.DebugEvent))

	expected = `Jan  2 15:04:00 ERROR error event: error message k1="some value" k2=2 k3=3.5 k4=true`
	checkRendered(t, expected, RenderString(HumanReadable, cuetest.ErrorEventNoFrames))

	expected = `Jan  2 15:04:00 ERROR file3.go:3 error event: error message k1="some value" k2=2 k3=3.5 k4=true`
	checkRendered(t, expected, RenderString(HumanReadable, cuetest.ErrorEvent))
}

func TestHumanReadableColors(t *testing.T) {
	expected := "\x1b[34mJan  2 15:04:00 DEBUG debug event k1=\"some value\" k2=2 k3=3.5 k4=true\x1b[0m"
	checkRendered(t, expected, RenderString(HumanReadableColors, cuetest.DebugEventNoFrames))

	expected = "\x1b[34mJan  2 15:04:00 DEBUG file3.go:3 debug event k1=\"some value\" k2=2 k3=3.5 k4=true\x1b[0m"
	checkRendered(t, expected, RenderString(HumanReadableColors, cuetest.DebugEvent))

	expected = "\x1b[31mJan  2 15:04:00 ERROR error event: error message k1=\"some value\" k2=2 k3=3.5 k4=true\x1b[0m"
	checkRendered(t, expected, RenderString(HumanReadableColors, cuetest.ErrorEventNoFrames))

	expected = "\x1b[31mJan  2 15:04:00 ERROR file3.go:3 error event: error message k1=\"some value\" k2=2 k3=3.5 k4=true\x1b[0m"
	checkRendered(t, expected, RenderString(HumanReadableColors, cuetest.ErrorEvent))
}

func TestJSONMessage(t *testing.T) {
	expected := `debug event {"k1":"some value","k2":2,"k3":3.5,"k4":true}`
	checkRendered(t, expected, RenderString(JSONMessage, cuetest.DebugEvent))

	expected = `error event: error message {"k1":"some value","k2":2,"k3":3.5,"k4":true}`
	checkRendered(t, expected, RenderString(JSONMessage, cuetest.ErrorEvent))
}

func TestJoin(t *testing.T) {
	checkRendered(t, "1 2 3", RenderString(Join(" ", Literal("1"), Literal("2"), Literal("3")), cuetest.DebugEvent))
	checkRendered(t, "1 3", RenderString(Join(" ", Literal("1"), Literal(""), Literal("3")), cuetest.DebugEvent))
	checkRendered(t, "1 2", RenderString(Join(" ", Literal("1"), Literal("2"), Literal("")), cuetest.DebugEvent))
	checkRendered(t, "2 3", RenderString(Join(" ", Literal(""), Literal("2"), Literal("3")), cuetest.DebugEvent))
}

func TestFormatf(t *testing.T) {
	checkRendered(t, "1 + 2 = 3", RenderString(Formatf("%v + %v = %v", Literal("1"), Literal("2"), Literal("3")), cuetest.DebugEvent))
	checkRendered(t, "1+2=3", RenderString(Formatf("%v+%v=%v", Literal("1"), Literal("2"), Literal("3")), cuetest.DebugEvent))
	checkRendered(t, " 1+2=3", RenderString(Formatf(" %v+%v=%v", Literal("1"), Literal("2"), Literal("3")), cuetest.DebugEvent))
	checkRendered(t, "1+2=3 ", RenderString(Formatf("%v+%v=%v ", Literal("1"), Literal("2"), Literal("3")), cuetest.DebugEvent))
	checkRendered(t, " 1+2=3 ", RenderString(Formatf(" %v+%v=%v ", Literal("1"), Literal("2"), Literal("3")), cuetest.DebugEvent))
	checkRendered(t, "test %v test", RenderString(Formatf("%v %%v %v", Literal("test"), Literal("test")), cuetest.DebugEvent))
	checkRendered(t, "test%vtest", RenderString(Formatf("%v%%v%v", Literal("test"), Literal("test")), cuetest.DebugEvent))
	checkRendered(t, " test%vtest", RenderString(Formatf(" %v%%v%v", Literal("test"), Literal("test")), cuetest.DebugEvent))
	checkRendered(t, "test%vtest ", RenderString(Formatf("%v%%v%v ", Literal("test"), Literal("test")), cuetest.DebugEvent))
	checkRendered(t, "test%v%vtest", RenderString(Formatf("%v%%v%%v%v", Literal("test"), Literal("test")), cuetest.DebugEvent))
	checkRendered(t, "test%%test", RenderString(Formatf("%v%%%%%v", Literal("test"), Literal("test")), cuetest.DebugEvent))
	checkRendered(t, "test %!v(MISSING)", RenderString(Formatf("test %v"), cuetest.DebugEvent))
}

func TestColorize(t *testing.T) {
	test := Literal("test")
	checkRendered(t, "\x1b[34mtest\x1b[0m", RenderString(Colorize(test), cuetest.DebugEvent))
	checkRendered(t, "\x1b[32mtest\x1b[0m", RenderString(Colorize(test), cuetest.InfoEvent))
	checkRendered(t, "\x1b[33mtest\x1b[0m", RenderString(Colorize(test), cuetest.WarnEvent))
	checkRendered(t, "\x1b[31mtest\x1b[0m", RenderString(Colorize(test), cuetest.ErrorEvent))
	checkRendered(t, "\x1b[31mtest\x1b[0m", RenderString(Colorize(test), cuetest.FatalEvent))
}

func TestTrim(t *testing.T) {
	checkRendered(t, "test", RenderString(Trim(Literal(" test ")), cuetest.DebugEvent))
	checkRendered(t, "test", RenderString(Trim(Literal("		test	")), cuetest.DebugEvent))
	checkRendered(t, "test", RenderString(Trim(Literal("\ttest\t")), cuetest.DebugEvent))
	checkRendered(t, "test", RenderString(Trim(Literal("\ntest\n")), cuetest.DebugEvent))
}

func TestEscape(t *testing.T) {
	checkRendered(t, "test", RenderString(Escape(Literal("test")), cuetest.DebugEvent))
	checkRendered(t, " test ", RenderString(Escape(Literal(" test ")), cuetest.DebugEvent))
	checkRendered(t, "日本", RenderString(Escape(Literal("日本")), cuetest.DebugEvent))
	checkRendered(t, "\\t", RenderString(Escape(Literal("\t")), cuetest.DebugEvent))
	checkRendered(t, "\\n", RenderString(Escape(Literal("\n")), cuetest.DebugEvent))
	checkRendered(t, "\\x00", RenderString(Escape(Literal("\x00")), cuetest.DebugEvent))
	checkRendered(t, "\\x00", RenderString(Escape(Literal(string(rune(0)))), cuetest.DebugEvent))
}

func TestTruncate(t *testing.T) {
	checkRendered(t, "tes", RenderString(Truncate(Literal("test"), 3), cuetest.DebugEvent))
}

func TestLiteral(t *testing.T) {
	checkRendered(t, "test", RenderString(Literal("test"), cuetest.DebugEvent))
}

func TestTime(t *testing.T) {
	checkRendered(t, "Jan  2 15:04:00", RenderString(Time(time.Stamp), cuetest.DebugEvent))
}

func TestHostname(t *testing.T) {
	host, err := os.Hostname()
	if err != nil {
		t.Errorf("Encountered unexpected error getting hostname: %s", err)
	}
	checkRendered(t, strings.Split(host, ".")[0], RenderString(Hostname, cuetest.DebugEvent))
}

func TestFQDN(t *testing.T) {
	host, err := os.Hostname()
	if err != nil {
		t.Errorf("Encountered unexpected error getting hostname: %s", err)
	}
	checkRendered(t, host, RenderString(FQDN, cuetest.DebugEvent))
}

func TestLevel(t *testing.T) {
	checkRendered(t, "DEBUG", RenderString(Level, cuetest.DebugEvent))
	checkRendered(t, "INFO", RenderString(Level, cuetest.InfoEvent))
	checkRendered(t, "WARN", RenderString(Level, cuetest.WarnEvent))
	checkRendered(t, "ERROR", RenderString(Level, cuetest.ErrorEvent))
	checkRendered(t, "FATAL", RenderString(Level, cuetest.FatalEvent))
}

func TestPackage(t *testing.T) {
	checkRendered(t, "github.com/bobziuchkovski/cue/frame3", RenderString(Package, cuetest.DebugEvent))
	checkRendered(t, cue.UnknownPackage, RenderString(Package, cuetest.DebugEventNoFrames))
}

func TestFunction(t *testing.T) {
	checkRendered(t, "github.com/bobziuchkovski/cue/frame3.function3", RenderString(Function, cuetest.DebugEvent))
	checkRendered(t, cue.UnknownFunction, RenderString(Function, cuetest.DebugEventNoFrames))
}

func TestFile(t *testing.T) {
	checkRendered(t, "/path/github.com/bobziuchkovski/cue/frame3/file3.go", RenderString(File, cuetest.DebugEvent))
	checkRendered(t, cue.UnknownFile, RenderString(File, cuetest.DebugEventNoFrames))
}

func TestShortFile(t *testing.T) {
	checkRendered(t, "file3.go", RenderString(ShortFile, cuetest.DebugEvent))
	checkRendered(t, cue.UnknownFile, RenderString(ShortFile, cuetest.DebugEventNoFrames))
}

func TestLine(t *testing.T) {
	checkRendered(t, "3", RenderString(Line, cuetest.DebugEvent))
	checkRendered(t, "0", RenderString(Line, cuetest.DebugEventNoFrames))
}

func TestMessage(t *testing.T) {
	checkRendered(t, "debug event", RenderString(Message, cuetest.DebugEvent))
	checkRendered(t, "error event", RenderString(Message, cuetest.ErrorEvent))
}

func TestError(t *testing.T) {
	checkRendered(t, "", RenderString(Error, cuetest.DebugEvent))
	checkRendered(t, "error message", RenderString(Error, cuetest.ErrorEvent))
}

func TestErrorType(t *testing.T) {
	checkRendered(t, "", RenderString(ErrorType, cuetest.DebugEvent))
	checkRendered(t, "errors.errorString", RenderString(ErrorType, cuetest.ErrorEvent))
}

func TestMessageWithError(t *testing.T) {
	checkRendered(t, "debug event", RenderString(MessageWithError, cuetest.DebugEvent))
	checkRendered(t, "error event: error message", RenderString(MessageWithError, cuetest.ErrorEvent))
}

func TestSourceWithLine(t *testing.T) {
	checkRendered(t, "file3.go:3", RenderString(SourceWithLine, cuetest.DebugEvent))
	checkRendered(t, "", RenderString(SourceWithLine, cuetest.DebugEventNoFrames))
}

func TestContextName(t *testing.T) {
	checkRendered(t, "test context", RenderString(ContextName, cuetest.DebugEvent))
}

func TestHumanContext(t *testing.T) {
	checkRendered(t, `k1="some value" k2=2 k3=3.5 k4=true`, RenderString(HumanContext, cuetest.DebugEvent))

	e := cuetest.GenerateEvent(cue.DEBUG, nil, "test", nil, 0)

	e.Context = cue.NewContext("empty value").WithValue("k1", "")
	checkRendered(t, `k1=""`, RenderString(HumanContext, e))

	e.Context = cue.NewContext("value with double quote").WithValue("k1", `test"test`)
	checkRendered(t, `k1="test\"test"`, RenderString(HumanContext, e))

	e.Context = cue.NewContext("value with single quote").WithValue("k1", `test'test`)
	checkRendered(t, `k1="test'test"`, RenderString(HumanContext, e))

	e.Context = cue.NewContext("value with backslash quote").WithValue("k1", `test\test`)
	checkRendered(t, `k1="test\\test"`, RenderString(HumanContext, e))

	e.Context = cue.NewContext("key with double quote").WithValue(`test"test`, "v1")
	checkRendered(t, `"test\"test"=v1`, RenderString(HumanContext, e))

	e.Context = cue.NewContext("key with single quote").WithValue(`test'test`, "v1")
	checkRendered(t, `"test'test"=v1`, RenderString(HumanContext, e))

	e.Context = cue.NewContext("key with backslash quote").WithValue(`test\test`, "v1")
	checkRendered(t, `"test\\test"=v1`, RenderString(HumanContext, e))

	e.Context = cue.NewContext("key and value needing quotes").WithValue(`test\test`, `v1 v2`)
	checkRendered(t, `"test\\test"="v1 v2"`, RenderString(HumanContext, e))
}

func TestJSONContext(t *testing.T) {
	checkRendered(t, `{"k1":"some value","k2":2,"k3":3.5,"k4":true}`, RenderString(JSONContext, cuetest.DebugEvent))
}

func TestStructuredContext(t *testing.T) {
	checkRendered(t, `k4="true" k3="3.5" k2="2" k1="some value"`, RenderString(StructuredContext, cuetest.DebugEvent))

	e := cuetest.GenerateEvent(cue.DEBUG, nil, "test", nil, 0)

	e.Context = cue.NewContext("invalid key 1").WithValue("k1", "v1").WithValue("日本", "country")
	checkRendered(t, `k1="v1"`, RenderString(StructuredContext, e))

	e.Context = cue.NewContext("invalid key 2").WithValue("k1", "v1").WithValue("k1=k1", "bad")
	checkRendered(t, `k1="v1"`, RenderString(StructuredContext, e))

	e.Context = cue.NewContext("invalid key 3").WithValue("k1", "v1").WithValue("k1]k1", "bad")
	checkRendered(t, `k1="v1"`, RenderString(StructuredContext, e))

	e.Context = cue.NewContext("invalid key 4").WithValue("k1", "v1").WithValue(`k1"k1`, "bad")
	checkRendered(t, `k1="v1"`, RenderString(StructuredContext, e))

	e.Context = cue.NewContext("invalid key 5").WithValue("k1", "v1").WithValue("k1\x00k1", "bad")
	checkRendered(t, `k1="v1"`, RenderString(StructuredContext, e))

	e.Context = cue.NewContext("invalid key 6").WithValue("k1", "v1").WithValue("really, really, super looooooooooooonnnnggggg key", "bad")
	checkRendered(t, `k1="v1"`, RenderString(StructuredContext, e))

	e.Context = cue.NewContext("escaped values").WithValue("k1", "v1").WithValue("escaped", `test ' test " test ] test \ test`)
	checkRendered(t, `escaped="test ' test \" test \] test \\ test" k1="v1"`, RenderString(StructuredContext, e))
}

func checkRendered(t *testing.T, expected string, result string) {
	if result != expected {
		t.Errorf("Expected to render %q, not %q", expected, result)
	}
}

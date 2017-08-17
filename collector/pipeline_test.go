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

package collector

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bobziuchkovski/cue"
	"github.com/bobziuchkovski/cue/internal/cuetest"
)

func TestPipelineContextFilter(t *testing.T) {
	c1 := cuetest.NewCapturingCollector()
	p1 := NewPipeline().FilterContext(func(key string, value interface{}) bool {
		return key == "k1"
	})
	p1.Attach(c1).Collect(cuetest.DebugEvent)

	fieldExpectation := cue.Fields{
		"k2": 2,
		"k3": 3.5,
		"k4": true,
	}
	if !reflect.DeepEqual(c1.Captured()[0].Context.Fields(), fieldExpectation) {
		t.Errorf("Expected to see altered context %v but saw %v instead", fieldExpectation, c1.Captured()[0].Context.Fields())
	}

	c2 := cuetest.NewCapturingCollector()
	p2 := NewPipeline().FilterContext(func(key string, value interface{}) bool {
		return key == "bogus"
	})
	p2.Attach(c2).Collect(cuetest.DebugEvent)

	if !reflect.DeepEqual(c2.Captured()[0].Context.Fields(), cuetest.DebugEvent.Context.Fields()) {
		t.Errorf("Expected to see an unaltered context, but saw %v instead", c2.Captured()[0].Context.Fields())
	}

	if c2.Captured()[0] == cuetest.DebugEvent {
		t.Error("Expected to see a cloned event, but saw our same input event instead")
	}
}

func TestPipelineEventFilter(t *testing.T) {
	c1 := cuetest.NewCapturingCollector()
	p1 := NewPipeline().FilterEvent(func(event *cue.Event) bool {
		return event.Level == cue.DEBUG
	})
	p1.Attach(c1).Collect(cuetest.DebugEvent)

	if len(c1.Captured()) != 0 {
		t.Errorf("Expected to see no events after filtering DEBUG level, but saw %d instead", len(c1.Captured()))
	}

	c2 := cuetest.NewCapturingCollector()
	p2 := NewPipeline().FilterEvent(func(event *cue.Event) bool {
		return event.Level == cue.ERROR
	})
	p2.Attach(c2).Collect(cuetest.DebugEvent)

	if len(c2.Captured()) != 1 {
		t.Errorf("Expected to a single event after filtering ERROR level, but saw %d instead", len(c2.Captured()))
	}

	if c2.Captured()[0] == cuetest.DebugEvent {
		t.Error("Expected to see a cloned event, but saw our same input event instead")
	}
}

func TestPipelineContextTransformer(t *testing.T) {
	c1 := cuetest.NewCapturingCollector()
	p1 := NewPipeline().TransformContext(func(ctx cue.Context) cue.Context {
		return cue.NewContext("replaced").WithValue("field", "value")
	})
	p1.Attach(c1).Collect(cuetest.DebugEvent)

	if len(c1.Captured()) != 1 {
		t.Errorf("Expected to see a single event but saw %d instead", len(c1.Captured()))
	}

	capturedCtx := c1.Captured()[0].Context
	if capturedCtx.Name() != "replaced" {
		t.Errorf("Expected to see context with name %q, not %q", "replaced", capturedCtx.Name())
	}

	if !reflect.DeepEqual(capturedCtx.Fields(), cue.Fields{"field": "value"}) {
		t.Errorf("Expected to see context values of %v, not %v", cue.Fields{"field": "value"}, capturedCtx.Fields())
	}
}

func TestPipelineEventTransformer(t *testing.T) {
	c1 := cuetest.NewCapturingCollector()
	p1 := NewPipeline().TransformEvent(func(event *cue.Event) *cue.Event {
		return cuetest.ErrorEvent
	})
	p1.Attach(c1).Collect(cuetest.DebugEvent)

	if len(c1.Captured()) != 1 {
		t.Errorf("Expected to see a single event but saw %d instead", len(c1.Captured()))
	}
	if !reflect.DeepEqual(cuetest.ErrorEvent, c1.Captured()[0]) {
		t.Error("Expected to see a copy of errorEVent, but didn't")
	}
}

func TestMultiPipeline(t *testing.T) {
	c1 := cuetest.NewCapturingCollector()
	p1 := NewPipeline().FilterContext(func(key string, value interface{}) bool {
		return key == "k1"
	}).FilterEvent(func(event *cue.Event) bool {
		return event.Level == cue.ERROR
	}).TransformEvent(func(event *cue.Event) *cue.Event {
		event.Message = "Replaced message"
		return event
	}).TransformContext(func(ctx cue.Context) cue.Context {
		return ctx.WithValue("addedkey", "addedvalue")
	})

	c2 := p1.Attach(c1)
	c2.Collect(cuetest.DebugEvent)
	c2.Collect(cuetest.ErrorEvent)
	c2.Collect(nil)
	cuetest.CloseCollector(c2)

	if len(c1.Captured()) != 1 {
		t.Errorf("Expected to see a single event but saw %d instead", len(c1.Captured()))
	}

	event := c1.Captured()[0]
	fieldExpectation := cue.Fields{
		"k2":       2,
		"k3":       3.5,
		"k4":       true,
		"addedkey": "addedvalue",
	}
	if !reflect.DeepEqual(event.Context.Fields(), fieldExpectation) {
		t.Errorf("Expected to see context fields of %v but saw %v instead", fieldExpectation, event.Context.Fields())
	}
	if event.Message != "Replaced message" {
		t.Errorf("Expected to see message content of %q not %q", "Replaced message", event.Message)
	}
}

func TestPipelineString(t *testing.T) {
	c1 := cuetest.NewCapturingCollector()
	p1 := NewPipeline().Attach(c1)

	// Ensure nothing panics
	_ = fmt.Sprint(p1)
}

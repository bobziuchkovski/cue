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
	"github.com/bobziuchkovski/cue"
	"io"
)

// ContextFilter is used with a Pipeline to filter context key/value pairs.
type ContextFilter func(key string, value interface{}) bool

// ContextTransformer is used with a Pipeline to transform context key/value
// pairs.
type ContextTransformer func(context cue.Context) cue.Context

// EventFilter is used with a Pipeline to filter events.
type EventFilter func(event *cue.Event) bool

// EventTransformer is used with a Pipeline to transform events.
type EventTransformer func(event *cue.Event) *cue.Event

// Pipeline is an immutable builder for Event and Context transforms.
// Pipeline methods create and return updated *Pipeline instances.  They are
// meant to be invoked as a chain.
//
// Hence the following is correct:
//
//		pipe := NewPipeline().FilterContext(...)
//		filtered := p.Attach(...)
//
// Whereas the following is incorrect and does nothing:
//
//		pipe := NewPipeline()
//		pipe.FilterContext(...)  // Wrong: the returned *Pipeline is ignored
//		filtered := p.Attach(...)
//
// Since pipeline objects are immutable, they may be attached to multiple
// collectors, and may be attached at multiple points during their build
// process to different collectors.
//
// Pipeline passes copies of input events to its filters/transformers, so the
// events may be modified in place.
type Pipeline struct {
	prior       *Pipeline
	transformer EventTransformer
}

// NewPipeline returns a new pipeline instance.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// FilterContext returns an updated copy of Pipeline that drops Context
// key/value pairs that match any of the provided filters.
func (p *Pipeline) FilterContext(filters ...ContextFilter) *Pipeline {
	return &Pipeline{
		prior:       p,
		transformer: filterNilEvent(filterContext(filters...)),
	}
}

// FilterEvent returns an updated copy of Pipeline that drops events
// that match any of the provided filters.
func (p *Pipeline) FilterEvent(filters ...EventFilter) *Pipeline {
	return &Pipeline{
		prior:       p,
		transformer: filterNilEvent(filterEvent(filters...)),
	}
}

// TransformContext returns an updated copy of Pipeline that transforms event
// contexts according to the provided transformers.
func (p *Pipeline) TransformContext(transformers ...ContextTransformer) *Pipeline {
	return &Pipeline{
		prior:       p,
		transformer: filterNilEvent(transformContext(transformers...)),
	}
}

// TransformEvent returns an updated copy of Pipeline that transforms events
// according to the provided transformers.
func (p *Pipeline) TransformEvent(transformers ...EventTransformer) *Pipeline {
	return &Pipeline{
		prior:       p,
		transformer: filterNilEvent(transformEvent(transformers...)),
	}
}

// Attach returns a new collector with the pipeline attached to c.
func (p *Pipeline) Attach(c cue.Collector) cue.Collector {
	if p.prior == nil {
		log.Warn("Pipeline.Attach called on an empty pipeline.")
	}
	return &pipelineCollector{
		pipeline:  p,
		collector: c,
	}
}

func (p *Pipeline) apply(event *cue.Event) *cue.Event {
	if event == nil {
		return nil
	}
	if p.prior == nil {
		return cloneEvent(event)
	}
	return p.transformer(p.prior.apply(event))
}

type pipelineCollector struct {
	pipeline  *Pipeline
	collector cue.Collector
}

func (p *pipelineCollector) String() string {
	return fmt.Sprintf("Pipeline(target=%s)", p.collector)
}

func (p *pipelineCollector) Collect(event *cue.Event) error {
	transformed := p.pipeline.apply(event)
	if transformed == nil {
		return nil
	}
	return p.collector.Collect(transformed)
}

func (p *pipelineCollector) Close() error {
	closer, ok := p.collector.(io.Closer)
	if !ok {
		return nil
	}
	return closer.Close()
}

func filterContext(filters ...ContextFilter) EventTransformer {
	return func(event *cue.Event) *cue.Event {
		newContext := cue.NewContext(event.Context.Name())
		event.Context.Each(func(key string, value interface{}) {
			for _, filter := range filters {
				if filter(key, value) {
					return
				}
			}
			newContext = newContext.WithValue(key, value)
		})
		event.Context = newContext
		return event
	}
}

func filterEvent(filters ...EventFilter) EventTransformer {
	return func(event *cue.Event) *cue.Event {
		for _, filter := range filters {
			if filter(event) {
				return nil
			}
		}
		return event
	}
}

func transformContext(transformers ...ContextTransformer) EventTransformer {
	return func(event *cue.Event) *cue.Event {
		for _, trans := range transformers {
			event.Context = trans(event.Context)
		}
		return event
	}
}

func transformEvent(transformers ...EventTransformer) EventTransformer {
	return func(event *cue.Event) *cue.Event {
		for _, trans := range transformers {
			if event == nil {
				return nil
			}
			event = trans(event)
		}
		return event
	}
}

func filterNilEvent(transformer EventTransformer) EventTransformer {
	return func(event *cue.Event) *cue.Event {
		if event == nil {
			return nil
		}
		return transformer(event)
	}
}

func cloneEvent(e *cue.Event) *cue.Event {
	return &cue.Event{
		Time:    e.Time,
		Level:   e.Level,
		Context: e.Context,
		Frames:  e.Frames,
		Error:   e.Error,
		Message: e.Message,
	}
}

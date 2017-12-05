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
	"fmt"
	"reflect"
)

var (
	emptyPairs = (*pairs)(nil)
	errorP     = (*error)(nil)
	errorT     = reflect.TypeOf(errorP).Elem()
	stringerP  = (*fmt.Stringer)(nil)
	stringerT  = reflect.TypeOf(stringerP).Elem()
)

// Fields is a map representation of contextual key/value pairs.
type Fields map[string]interface{}

// Context is an interface representing contextual key/value pairs.  Any
// key/value pair may be added to a context with one exception: an empty string
// is not a valid key.  Pointer values are dereferenced and their target is
// added.  Values of basic types -- string, bool, integer, float, and complex
// -- are stored directly.  Other types, including all slices and arrays, are
// coerced to a string representation via fmt.Sprint.  This ensures stored
// context values are immutable.  This is important for safe asynchronous
// operation.
//
// Storing duplicate keys is allowed, but the resulting behavior is currently
// undefined.
type Context interface {
	// Name returns the name of the context.
	Name() string

	// NumValues returns the number of key/value pairs in the Context.
	// The counting behavior for duplicate keys is currently undefined.
	NumValues() int

	// Each executes function fn on each of the Context's key/value pairs.
	// Iteration order is currently undefined.
	Each(fn func(key string, value interface{}))

	// Fields returns a map representation of the Context's key/value pairs.
	// Duplicate key handling is currently undefined.
	Fields() Fields

	// WithFields returns a new Context that adds the key/value pairs from
	// fields to the existing key/value pairs.
	WithFields(fields Fields) Context

	// WithValue returns a new Context that adds key and value to the existing
	// key/value pairs.
	WithValue(key string, value interface{}) Context
}

type context struct {
	name  string
	pairs *pairs
}

// JoinContext returns a new Context with the given name, containing all the
// key/value pairs joined from the provided contexts.
func JoinContext(name string, contexts ...Context) Context {
	// This is pretty inefficient...we could probably create a wrapper view
	// that dispatches to the underlying contexts if needed.
	joined := NewContext(name)
	for _, context := range contexts {
		if context == nil {
			continue
		}
		context.Each(func(key string, value interface{}) {
			joined = joined.WithValue(key, value)
		})
	}
	return joined
}

// NewContext returns a new Context with the given name.
func NewContext(name string) Context {
	return &context{
		name:  name,
		pairs: emptyPairs,
	}
}

func (c *context) String() string {
	return fmt.Sprintf("Context(name=%s)", c.name)
}

func (c *context) Name() string {
	return c.name
}

func (c *context) NumValues() int {
	return c.pairs.count()
}

func (c *context) Each(fn func(key string, value interface{})) {
	c.pairs.each(fn)
}

func (c *context) Fields() Fields {
	return c.pairs.toFields()
}

func (c *context) WithFields(fields Fields) Context {
	var new Context = c
	for k, v := range fields {
		new = new.WithValue(k, v)
	}
	return new
}

func (c *context) WithValue(key string, value interface{}) Context {
	if key == "" {
		return c
	}
	return &context{
		name:  c.name,
		pairs: c.pairs.append(key, basicValue(value)),
	}
}

type pairs struct {
	prev  *pairs
	key   string
	value interface{}
}

func (p *pairs) append(key string, value interface{}) *pairs {
	return &pairs{
		prev:  p,
		key:   key,
		value: value,
	}
}

func (p *pairs) each(fn func(key string, value interface{})) {
	for current := p; current != nil; current = current.prev {
		fn(current.key, current.value)
	}
}

func (p *pairs) count() int {
	count := 0
	for current := p; current != nil; current = current.prev {
		count++
	}
	return count
}

func (p *pairs) toFields() Fields {
	if p == nil {
		return make(Fields)
	}
	fields := p.prev.toFields()
	fields[p.key] = p.value
	return fields
}

// basicValue serves to dereference pointers and coerce non-basic types to strings,
// ensuring all values are effectively immutable.  The latter is critical for
// asynchronous operation.  We can't have context values changing while an event is
// queued, or else the logged value won't represent the value as it was at the
// time the event was generated.
func basicValue(value interface{}) interface{} {
	rval := reflect.ValueOf(value)
	if !rval.IsValid() {
		return fmt.Sprint(value)
	}
	for rval.Kind() == reflect.Ptr {
		if rval.IsNil() {
			break
		}
		if rval.Type().Implements(stringerT) || rval.Type().Implements(errorT) {
			break
		}
		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.Bool, reflect.String:
		return rval.Interface()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rval.Interface()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rval.Interface()
	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return rval.Interface()
	default:
		return fmt.Sprint(rval.Interface())
	}
}

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
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"testing"
)

var contextFieldTests = []struct {
	Name       string
	Context    Context
	Logger     Logger
	FieldEquiv Fields
}{
	{
		Name:       "Empty",
		Context:    NewContext("Empty"),
		Logger:     NewLogger("Empty"),
		FieldEquiv: Fields{},
	},
	{
		Name:       "WithValue",
		Context:    NewContext("WithValue").WithValue("k1", "v1"),
		Logger:     NewLogger("WithValue").WithValue("k1", "v1"),
		FieldEquiv: Fields{"k1": "v1"},
	},
	{
		Name:       "WithFields",
		Context:    NewContext("WithFields").WithFields(Fields{"k1": "v1", "k2": 2}),
		Logger:     NewLogger("WithFields").WithFields(Fields{"k1": "v1", "k2": 2}),
		FieldEquiv: Fields{"k1": "v1", "k2": 2},
	},
	{
		Name:       "Chained1",
		Context:    NewContext("Chained1").WithFields(Fields{"k1": "v1", "k2": 2}).WithValue("k3", 3.0),
		Logger:     NewLogger("Chained1").WithFields(Fields{"k1": "v1", "k2": 2}).WithValue("k3", 3.0),
		FieldEquiv: Fields{"k1": "v1", "k2": 2, "k3": 3.0},
	},
	{
		Name:       "Chained2",
		Context:    NewContext("Chained2").WithValue("k1", "v1").WithFields(Fields{"k2": 2, "k3": 3.0}),
		Logger:     NewLogger("Chained2").WithValue("k1", "v1").WithFields(Fields{"k2": 2, "k3": 3.0}),
		FieldEquiv: Fields{"k1": "v1", "k2": 2, "k3": 3.0},
	},
}

func TestContextName(t *testing.T) {
	c1 := NewContext("test")
	if c1.Name() != "test" {
		t.Errorf("Expected %q for Context Name, but received %q instead", "test", c1.Name())
	}
}

func TestContextNumValues(t *testing.T) {
	for _, test := range contextFieldTests {
		num := test.Context.NumValues()
		if num != len(test.FieldEquiv) {
			t.Errorf("Incorrect count for NumValues().  Test: %s, Expected: %d, Received: %d", test.Name, len(test.FieldEquiv), num)
		}
	}
}

func TestContextEach(t *testing.T) {
	for _, test := range contextFieldTests {
		visited := make(Fields)
		test.Context.Each(func(key string, value interface{}) {
			visited[key] = value
		})

		if !reflect.DeepEqual(visited, test.FieldEquiv) {
			t.Errorf("Visited fields don't match.  Test: %s, Expected: %#v, Received: %#v", test.Name, test.FieldEquiv, visited)
		}
	}
}

func TestContextFields(t *testing.T) {
	for _, test := range contextFieldTests {
		if !reflect.DeepEqual(test.Context.Fields(), test.FieldEquiv) {
			t.Errorf("Visited fields don't match.  Test: %s, Expected: %#v, Received: %#v", test.Name, test.FieldEquiv, test.Context.Fields())
		}
	}
}

func TestContextString(t *testing.T) {
	c := NewContext("test")
	s, ok := c.(fmt.Stringer)
	if !ok {
		t.Error("Expected context type to implement String() method but it doesn't")
	}
	_ = s.String()
}

func TestContextEmptyKey(t *testing.T) {
	c1 := NewContext("test")
	c2 := c1.WithValue("", "empty key")
	if c1 != c2 {
		t.Error("Expected WithValue to return identity if key is empty")
	}
}

func TestJoinContext(t *testing.T) {
	c1 := NewContext("first").WithValue("k1", "v1").WithFields(Fields{"k2": 2, "k3": 3.0})
	c2 := NewContext("second").WithFields(Fields{"k4": "v4", "k5": true}).WithValue("k6", uintptr(0x12345678))
	joined := JoinContext("joined", c1, c2)
	if joined.Name() != "joined" {
		t.Errorf("Context name is incorrect.  Expected: %q, Received: %q", "joined", joined.Name())
	}
	expected := Fields{
		"k1": "v1",
		"k2": 2,
		"k3": 3.0,
		"k4": "v4",
		"k5": true,
		"k6": uintptr(0x12345678),
	}
	if !reflect.DeepEqual(joined.Fields(), expected) {
		t.Errorf("Context values are incorrect.  Expected: %v, Received: %v", expected, joined.Fields())
	}
}

func TestJoinNilContext(t *testing.T) {
	joined := JoinContext("joined", nil, nil)
	if joined.Name() != "joined" {
		t.Errorf("Context name is incorrect.  Expected: %q, Received: %q", "joined", joined.Name())
	}
	if joined.NumValues() != 0 {
		t.Errorf("Expected to see 0 values but saw %d instead", joined.NumValues())
	}
}

var boolValue = true
var boolValuePtr = &boolValue
var boolValuePtrPtr = &boolValuePtr
var intValue = int(math.MinInt32)
var intValuePtr = &intValue
var intValuePtrPtr = &intValuePtr
var int8Value = math.MinInt8
var int8ValuePtr = &int8Value
var int8ValuePtrPtr = &int8ValuePtr
var int16Value = math.MinInt16
var int16ValuePtr = &int16Value
var int16ValuePtrPtr = &int16ValuePtr
var int32Value = math.MinInt32
var int32ValuePtr = &int32Value
var int32ValuePtrPtr = &int32ValuePtr
var int64Value = math.MinInt64
var int64ValuePtr = &int64Value
var int64ValuePtrPtr = &int64ValuePtr
var uintValue = uint(math.MaxUint32)
var uintValuePtr = &uintValue
var uintValuePtrPtr = &uintValuePtr
var uint8Value = math.MaxUint8
var uint8ValuePtr = &uint8Value
var uint8ValuePtrPtr = &uint8ValuePtr
var uint16Value = math.MaxUint16
var uint16ValuePtr = &uint16Value
var uint16ValuePtrPtr = &uint16ValuePtr
var uint32Value = math.MaxUint32
var uint32ValuePtr = &uint32Value
var uint32ValuePtrPtr = &uint32ValuePtr
var uint64Value = uint64(math.MaxUint64)
var uint64ValuePtr = &uint64Value
var uint64ValuePtrPtr = &uint64ValuePtr
var uintptrValue = uintptr(0x12345678)
var uintptrValuePtr = &uintptrValue
var uintptrValuePtrPtr = &uintptrValuePtr
var float32Value = math.MaxFloat32
var float32ValuePtr = &float32Value
var float32ValuePtrPtr = &float32ValuePtr
var float64Value = math.MaxFloat64
var float64ValuePtr = &float64Value
var float64ValuePtrPtr = &float64ValuePtr
var complex64Value = complex64(-5.0 + 12i)
var complex64ValuePtr = &complex64Value
var complex64ValuePtrPtr = &complex64ValuePtr
var complex128Value = complex128(-12.0 + 24i)
var complex128ValuePtr = &complex128Value
var complex128ValuePtrPtr = &complex128ValuePtr
var stringValue = "test string"
var stringValuePtr = &stringValue
var stringValuePtrPtr = &stringValuePtr
var arrayValue = [2]int{1, 2}
var arrayValuePtr = &arrayValue
var arrayValuePtrPtr = &arrayValuePtr
var chanValue = make(chan int)
var chanValuePtr = &chanValue
var chanValuePtrPtr = &chanValuePtr
var funcValue = func(v int) {}
var funcValuePtr = &funcValue
var funcValuePtrPtr = &funcValuePtr
var mapValue = map[int]string{1: "1", 2: "2"}
var mapValuePtr = &mapValue
var mapValuePtrPtr = &mapValuePtr
var sliceValue = []int{1, 2, 3}
var sliceValuePtr = &sliceValue
var sliceValuePtrPtr = &sliceValuePtr
var structValue = struct{ Value int }{Value: 1}
var structValuePtr = &structValue
var structValuePtrPtr = &structValuePtr
var iface io.Reader = bytes.NewReader(nil)
var ifacePtr = &iface
var ifacePtrPtr = &ifacePtr
var nilIface = (fmt.Stringer)(nil)
var nilIfacePtr = &nilIface
var nilIfacePtrPtr = &nilIfacePtr
var errIface = errors.New("error interface")
var errIfacePtr = &errIface
var errIfacePtrPtr = &errIfacePtr
var stringerIface = stringer{val: "stringer interface"}
var stringerIfacePtr = &stringerIface
var stringerIfacePtrPtr = &stringerIfacePtr
var nilPtr = (*int)(nil)
var nilPtrPtr = &nilPtr

type stringer struct{ val string }

func (s stringer) String() string { return s.val }

var contextValueTests = []struct {
	Name                       string
	Input                      interface{}
	Captured                   interface{}
	NonDeterministicFormatting bool
}{
	{
		Name:     "bool",
		Input:    boolValue,
		Captured: boolValue,
	},
	{
		Name:     "pointer to bool",
		Input:    boolValuePtr,
		Captured: boolValue,
	},
	{
		Name:     "pointer to pointer to bool",
		Input:    boolValuePtrPtr,
		Captured: boolValue,
	},
	{
		Name:     "int",
		Input:    intValue,
		Captured: intValue,
	},
	{
		Name:     "pointer to int",
		Input:    intValuePtr,
		Captured: intValue,
	},
	{
		Name:     "pointer to pointer to int",
		Input:    intValuePtrPtr,
		Captured: intValue,
	},
	{
		Name:     "int8",
		Input:    int8Value,
		Captured: int8Value,
	},
	{
		Name:     "pointer to int8",
		Input:    int8ValuePtr,
		Captured: int8Value,
	},
	{
		Name:     "pointer to pointer to int8",
		Input:    int8ValuePtrPtr,
		Captured: int8Value,
	},
	{
		Name:     "int16",
		Input:    int16Value,
		Captured: int16Value,
	},
	{
		Name:     "pointer to int16",
		Input:    int16ValuePtr,
		Captured: int16Value,
	},
	{
		Name:     "pointer to pointer to int16",
		Input:    int16ValuePtrPtr,
		Captured: int16Value,
	},
	{
		Name:     "int32",
		Input:    int32Value,
		Captured: int32Value,
	},
	{
		Name:     "pointer to int32",
		Input:    int32ValuePtr,
		Captured: int32Value,
	},
	{
		Name:     "pointer to pointer to int32",
		Input:    int32ValuePtrPtr,
		Captured: int32Value,
	},
	{
		Name:     "int64",
		Input:    int64Value,
		Captured: int64Value,
	},
	{
		Name:     "pointer to int64",
		Input:    int64ValuePtr,
		Captured: int64Value,
	},
	{
		Name:     "pointer to pointer to int64",
		Input:    int64ValuePtrPtr,
		Captured: int64Value,
	},
	{
		Name:     "uint",
		Input:    uintValue,
		Captured: uintValue,
	},
	{
		Name:     "pointer to uint",
		Input:    uintValuePtr,
		Captured: uintValue,
	},
	{
		Name:     "pointer to pointer to uint",
		Input:    uintValuePtrPtr,
		Captured: uintValue,
	},
	{
		Name:     "uint8",
		Input:    uint8Value,
		Captured: uint8Value,
	},
	{
		Name:     "pointer to uint8",
		Input:    uint8ValuePtr,
		Captured: uint8Value,
	},
	{
		Name:     "pointer to pointer to uint8",
		Input:    uint8ValuePtrPtr,
		Captured: uint8Value,
	},
	{
		Name:     "uint16",
		Input:    uint16Value,
		Captured: uint16Value,
	},
	{
		Name:     "pointer to uint16",
		Input:    uint16ValuePtr,
		Captured: uint16Value,
	},
	{
		Name:     "pointer to pointer to uint16",
		Input:    uint16ValuePtrPtr,
		Captured: uint16Value,
	},
	{
		Name:     "uint32",
		Input:    uint32Value,
		Captured: uint32Value,
	},
	{
		Name:     "pointer to uint32",
		Input:    uint32ValuePtr,
		Captured: uint32Value,
	},
	{
		Name:     "pointer to pointer to uint32",
		Input:    uint32ValuePtrPtr,
		Captured: uint32Value,
	},
	{
		Name:     "uint64",
		Input:    uint64Value,
		Captured: uint64Value,
	},
	{
		Name:     "pointer to uint64",
		Input:    uint64ValuePtr,
		Captured: uint64Value,
	},
	{
		Name:     "pointer to pointer to uint64",
		Input:    uint64ValuePtrPtr,
		Captured: uint64Value,
	},
	{
		Name:     "uintptr",
		Input:    uintptrValue,
		Captured: uintptrValue,
	},
	{
		Name:     "pointer to uintptr",
		Input:    uintptrValuePtr,
		Captured: uintptrValue,
	},
	{
		Name:     "pointer to pointer to uintptr",
		Input:    uintptrValuePtrPtr,
		Captured: uintptrValue,
	},
	{
		Name:     "float32",
		Input:    float32Value,
		Captured: float32Value,
	},
	{
		Name:     "pointer to float32",
		Input:    float32ValuePtr,
		Captured: float32Value,
	},
	{
		Name:     "pointer to pointer to float32",
		Input:    float32ValuePtrPtr,
		Captured: float32Value,
	},
	{
		Name:     "float64",
		Input:    float64Value,
		Captured: float64Value,
	},
	{
		Name:     "pointer to float64",
		Input:    float64ValuePtr,
		Captured: float64Value,
	},
	{
		Name:     "pointer to pointer to float64",
		Input:    float64ValuePtrPtr,
		Captured: float64Value,
	},
	{
		Name:     "complex64",
		Input:    complex64Value,
		Captured: complex64Value,
	},
	{
		Name:     "pointer to complex64",
		Input:    complex64ValuePtr,
		Captured: complex64Value,
	},
	{
		Name:     "pointer to pointer to complex64",
		Input:    complex64ValuePtrPtr,
		Captured: complex64Value,
	},
	{
		Name:     "complex128",
		Input:    complex128Value,
		Captured: complex128Value,
	},
	{
		Name:     "pointer to complex128",
		Input:    complex128ValuePtr,
		Captured: complex128Value,
	},
	{
		Name:     "pointer to pointer to complex128",
		Input:    complex128ValuePtrPtr,
		Captured: complex128Value,
	},
	{
		Name:     "string",
		Input:    stringValue,
		Captured: stringValue,
	},
	{
		Name:     "pointer to string",
		Input:    stringValuePtr,
		Captured: stringValue,
	},
	{
		Name:     "array",
		Input:    arrayValue,
		Captured: fmt.Sprint(arrayValue),
	},
	{
		Name:     "pointer to array",
		Input:    arrayValuePtr,
		Captured: fmt.Sprint(arrayValue),
	},
	{
		Name:     "pointer to pointer to array",
		Input:    arrayValuePtrPtr,
		Captured: fmt.Sprint(arrayValue),
	},
	{
		Name:     "chan",
		Input:    chanValue,
		Captured: fmt.Sprint(chanValue),
	},
	{
		Name:     "pointer to chan",
		Input:    chanValuePtr,
		Captured: fmt.Sprint(chanValue),
	},
	{
		Name:     "pointer to pointer to chan",
		Input:    chanValuePtrPtr,
		Captured: fmt.Sprint(chanValue),
	},
	// The next three tests cause go vet to choke.  I'm not sure how to explicitly
	// whitelist them, so for now they're commented.  I value the ability to use
	// go vet on CI builds more than testing the formatting of function values.
	/*
		{
			Name:     "func",
			Input:    funcValue,
			Captured: fmt.Sprint(funcValue),
		},
		{
			Name:     "pointer to func",
			Input:    funcValuePtr,
			Captured: fmt.Sprint(funcValue),
		},
		{
			Name:     "pointer to pointer to func",
			Input:    funcValuePtrPtr,
			Captured: fmt.Sprint(funcValue),
		},
	*/
	{
		Name:                       "map",
		Input:                      mapValue,
		Captured:                   fmt.Sprint(mapValue),
		NonDeterministicFormatting: true,
	},
	{
		Name:                       "pointer to map",
		Input:                      mapValuePtr,
		Captured:                   fmt.Sprint(mapValue),
		NonDeterministicFormatting: true,
	},
	{
		Name:                       "pointer to pointer to map",
		Input:                      mapValuePtrPtr,
		Captured:                   fmt.Sprint(mapValue),
		NonDeterministicFormatting: true,
	},
	{
		Name:     "slice",
		Input:    sliceValue,
		Captured: fmt.Sprint(sliceValue),
	},
	{
		Name:     "pointer to slice",
		Input:    sliceValuePtr,
		Captured: fmt.Sprint(sliceValue),
	},
	{
		Name:     "pointer to pointer to slice",
		Input:    sliceValuePtrPtr,
		Captured: fmt.Sprint(sliceValue),
	},
	{
		Name:     "struct",
		Input:    structValue,
		Captured: fmt.Sprint(structValue),
	},
	{
		Name:     "pointer to struct",
		Input:    structValuePtr,
		Captured: fmt.Sprint(structValue),
	},
	{
		Name:     "pointer to pointer to struct",
		Input:    structValuePtrPtr,
		Captured: fmt.Sprint(structValue),
	},
	{
		Name:                       "interface",
		Input:                      iface,
		Captured:                   fmt.Sprint(iface),
		NonDeterministicFormatting: true,
	},
	{
		Name:     "pointer to interface",
		Input:    ifacePtr,
		Captured: fmt.Sprint(iface),
	},
	{
		Name:     "pointer to pointer to interface",
		Input:    ifacePtrPtr,
		Captured: fmt.Sprint(iface),
	},
	{
		Name:     "nil interface",
		Input:    nilIface,
		Captured: fmt.Sprint(nilIface),
	},
	{
		Name:     "pointer to nil interface",
		Input:    nilIfacePtr,
		Captured: fmt.Sprint(nilIface),
	},
	{
		Name:     "pointer to pointer to nil interface",
		Input:    nilIfacePtrPtr,
		Captured: fmt.Sprint(nilIface),
	},
	{
		Name:     "error interface",
		Input:    errIface,
		Captured: errIface.Error(),
	},
	{
		Name:     "pointer to error interface",
		Input:    errIfacePtr,
		Captured: errIface.Error(),
	},
	{
		Name:     "pointer to pointer to error interface",
		Input:    errIfacePtrPtr,
		Captured: errIface.Error(),
	},
	{
		Name:     "stringer interface",
		Input:    stringerIface,
		Captured: stringerIface.String(),
	},
	{
		Name:     "pointer to stringer interface",
		Input:    stringerIfacePtr,
		Captured: stringerIface.String(),
	},
	{
		Name:     "pointer to pointer to stringer interface",
		Input:    stringerIfacePtrPtr,
		Captured: stringerIface.String(),
	},
	{
		Name:     "nil pointer",
		Input:    nilPtr,
		Captured: "<nil>",
	},
	{
		Name:     "pointer to nil pointer",
		Input:    nilPtrPtr,
		Captured: "<nil>",
	},
}

func TestContextValueCapture(t *testing.T) {
	for _, test := range contextValueTests {
		ctx := NewContext(test.Name).WithValue("value", test.Input)
		captured := ctx.Fields()["value"]
		if test.NonDeterministicFormatting {
			if reflect.TypeOf(captured) != reflect.TypeOf(test.Captured) {
				t.Errorf("Captured value type is incorrect.  Test: %s, Expected type: %s, Received type: %s", test.Name, reflect.TypeOf(test.Captured), reflect.TypeOf(captured))
			}
		} else {
			if captured != test.Captured {
				fmt.Printf("test.Captured type: %s, captured type: %s\n", reflect.TypeOf(test.Captured), reflect.TypeOf(captured))
				t.Errorf("Captured value is incorrect.  Test: %s, Expected: %v, Received: %v", test.Name, test.Captured, captured)
			}
		}
	}
}

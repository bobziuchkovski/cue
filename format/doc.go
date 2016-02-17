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
Package format implements event formatting.

Default Formatters

The HumanMessage and HumanReadable formats are used as default formatters for
most of the cue/collector and cue/hosted implementations.  These are a good
place to start, both for selecting a formatter, and for understanding how to
implement custom formats.

Custom Formatting

While the Formatter interface is easy to implement, it's simpler to assemble a
format using the existing formatting functions as building blocks.  The Join
and Formatf functions are particularly useful in this regard.  Both assemble
a new Formatter based on input formatters.  See the predefined formats for
examples.

Buffers

All formatters append to a Buffer.  The interface is similar to a bytes.Buffer,
but with a simpler API.  See the Buffer type documentation for details.
*/
package format

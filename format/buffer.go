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
	"errors"
	"sync"
	"unicode/utf8"
)

var (
	errBarrier = errors.New("cue/format: write attempted on a buffer that was previously read")
	pool       = newPool()
)

// Using a buffer pool brought basic benchmark runs down from 400016 ns/op to
// 3306 ns/op with a simple test that collected log.Info("test") to a collector
// that applied the HumanReadable format.
type bufferPool struct {
	pool *sync.Pool
}

func newPool() *bufferPool {
	return &bufferPool{pool: &sync.Pool{
		New: func() interface{} {
			return newBuffer()
		},
	}}
}

func (p *bufferPool) get() Buffer {
	buffer := p.pool.Get().(Buffer)
	buffer.Reset()
	return buffer
}

func (p *bufferPool) put(b Buffer) {
	p.pool.Put(b)
}

// Buffer represents a simple byte buffer.  It's similar to bytes.Buffer but
// with a simpler API and implemented as an interface.
type Buffer interface {
	// Bytes returns the buffered bytes.
	Bytes() []byte

	// Len Returns the number of buffered bytes.
	Len() int

	// Reset restores the buffer to a blank/empty state.  The underlying byte
	// slice is retained.
	Reset()

	// Write appends the byte slice value to the buffer.
	Write(value []byte)

	// WriteByte appends the byte value to the buffer.
	WriteByte(value byte)

	// WriteRune appends the rune value to the buffer.
	WriteRune(value rune)

	// WriteString appends the string value to the buffer.
	WriteString(value string)
}

type buffer struct {
	bytes   []byte
	runebuf [utf8.MaxRune]byte
}

// GetBuffer returns an empty buffer from a pool of Buffers.  A corresponding
// "defer ReleaseBuffer()" should be used to free the buffer when finished.
func GetBuffer() Buffer {
	return pool.get()
}

// ReleaseBuffer returns a buffer to the buffer pool.  Failing to release the
// buffer won't cause any harm, as the Go runtime will garbage collect it.
// However, as of Go 1.6, there's a significant performance gain in pooling and
// reusing Buffer instances.
func ReleaseBuffer(buffer Buffer) {
	pool.put(buffer)
}

// newBuffer creates a new buffer instance.  Currently, the initialized
// capacity is 64 bytes, but this may change.  The buffer grows automatically
// as needed.
func newBuffer() Buffer {
	return &buffer{
		bytes: make([]byte, 0, 64),
	}
}

func (b *buffer) Reset() {
	b.bytes = b.bytes[:0]
}

func (b *buffer) Bytes() []byte {
	return b.bytes
}

func (b *buffer) Len() int {
	return len(b.bytes)
}

func (b *buffer) WriteByte(value byte) {
	b.ensureCapacity(1)
	b.bytes = append(b.bytes, value)
}

func (b *buffer) WriteRune(value rune) {
	if value < utf8.RuneSelf {
		b.WriteByte(byte(value))
		return
	}
	size := utf8.EncodeRune(b.runebuf[:], value)
	b.Write(b.runebuf[:size])
}

func (b *buffer) WriteString(value string) {
	origlen := len(b.bytes)
	b.ensureCapacity(len(value))
	b.bytes = b.bytes[:origlen+len(value)]
	copy(b.bytes[origlen:], value)
}

func (b *buffer) Write(value []byte) {
	origlen := len(b.bytes)
	b.ensureCapacity(len(value))
	b.bytes = b.bytes[:origlen+len(value)]
	copy(b.bytes[origlen:], value)
}

func (b *buffer) ensureCapacity(size int) {
	curlen := len(b.bytes)
	curcap := cap(b.bytes)
	if curlen+size > curcap {
		new := make([]byte, curlen, 2*curcap+size)
		copy(new, b.bytes)
		b.bytes = new
	}
}

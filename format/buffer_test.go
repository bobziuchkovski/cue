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
	"strings"
	"testing"
)

func TestEmptyBuffer(t *testing.T) {
	buf := newBuffer()
	result := buf.Bytes()
	if len(result) != 0 {
		t.Errorf("Expected length of new buffer to be 0, not %d", len(result))
	}
}

func TestBufferWrite(t *testing.T) {
	buf := newBuffer()
	var slice []byte
	for i := 0; i < 255; i++ {
		slice = append(slice, byte(i))
	}
	buf.Write(slice)

	result := buf.Bytes()
	if len(result) != 255 {
		t.Errorf("Expected 255 bytes written, but saw %d instead", len(result))
	}
	for i, b := range result {
		if b != byte(i) {
			t.Errorf("Expected byte at offset %d to be 0x%02x but got 0x%02x instead", i, byte(i), b)
		}
	}
}

func TestBufferWriteString(t *testing.T) {
	buf := newBuffer()
	buf.WriteString("hello")
	buf.WriteString(" ")
	buf.WriteString("world")
	if string(buf.Bytes()) != "hello world" {
		t.Errorf("Expected buffer contents to be %q, not %q", "hello world", string(buf.Bytes()))
	}

	buf = newBuffer()
	longstr := strings.Repeat("hello", 1000)
	buf.WriteString(longstr)
	if string(buf.Bytes()) != longstr {
		t.Errorf("Expected buffer contents to be %q, not %q", longstr, string(buf.Bytes()))
	}
}

func TestBufferWriteRune(t *testing.T) {
	buf := newBuffer()
	buf.WriteRune('日')
	buf.WriteRune('本')
	if string(buf.Bytes()) != "日本" {
		t.Errorf("Expected buffer contents to be %q, not %q", "hello", string(buf.Bytes()))
	}

	buf = newBuffer()
	for i := 0; i < 1000; i++ {
		buf.WriteRune('h')
		buf.WriteRune('e')
		buf.WriteRune('l')
		buf.WriteRune('l')
		buf.WriteRune('o')
	}
	longstr := strings.Repeat("hello", 1000)
	if string(buf.Bytes()) != longstr {
		t.Errorf("Expected buffer contents to be %q, not %q", "hello", string(buf.Bytes()))
	}
}

func TestWriteByte(t *testing.T) {
	buf := newBuffer()
	for i := 0; i < 255; i++ {
		buf.WriteByte(byte(i))
	}

	result := buf.Bytes()
	if len(result) != 255 {
		t.Errorf("Expected 255 bytes written, but saw %d instead", len(result))
	}
	for i, b := range result {
		if b != byte(i) {
			t.Errorf("Expected byte at offset %d to be 0x%02x but got 0x%02x instead", i, byte(i), b)
		}
	}
}

func TestBufferLen(t *testing.T) {
	buf := newBuffer()
	for i := 0; i < 255; i++ {
		if buf.Len() != i {
			t.Errorf("Expected length to equal i (%d), not %d", i, buf.Len())
		}
		buf.Write([]byte{byte(i)})
	}
}

func TestBufferReset(t *testing.T) {
	buf := newBuffer()
	buf.WriteString("test")
	buf.Reset()
	if buf.Len() != 0 {
		t.Errorf("Buffer should be 0 after reset, but it's %d instead", buf.Len())
	}
	if len(buf.Bytes()) != 0 {
		t.Errorf("Buffer.Bytes() should have 0 length after reset, but it's %d instead", len(buf.Bytes()))
	}
	if cap(buf.Bytes()) < 4 {
		t.Errorf("Buffer.Bytes() should have capacity greater than or equal to the size of the original buffer (%d), but it's %d instead", len("test"), cap(buf.Bytes()))
	}
}

func TestGetBuffer(t *testing.T) {
	buf := GetBuffer()
	if buf.Len() != 0 {
		t.Errorf("GetBuffer should return a 0 length buffer, but it's %d length instead", buf.Len())
	}
	if len(buf.Bytes()) != 0 {
		t.Errorf("GetBuffer should return a 0 length buffer, but it's %d instead", len(buf.Bytes()))
	}
}

func TestReleaseBuffer(t *testing.T) {
	// Basic test to ensure the release doesn't panic
	buf := GetBuffer()
	ReleaseBuffer(buf)
}

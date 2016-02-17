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

package hosted

import (
	"encoding/hex"
	"testing"
)

func TestUUID(t *testing.T) {
	value := hex.EncodeToString(uuid())
	if value[12] != byte('4') {
		t.Errorf("Invalid UUID.  Expected the 13th character to be '4' but got %q instead.  UUID: %s", rune(value[13]), value)
	}
	switch value[16] {
	case byte('8'), byte('9'), byte('a'), byte('b'):
		// Valid
	default:
		t.Errorf("Invalid UUID.  Expected the 17th character to be '8', '9', 'a', or 'b', but got %q instead.  UUID: %s", rune(value[16]), value)
	}
}

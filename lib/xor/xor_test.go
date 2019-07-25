// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package xor

import (
	"bytes"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	orig := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	enc := Encode(orig)
	dec := Decode(enc)
	if !bytes.Equal(orig, dec) {
		t.Fail()
	}
}

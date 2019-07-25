// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package xor

func Decode(b []byte) []byte {
	if len(b) < 1 {
		return []byte{}
	}
	out := make([]byte, len(b))
	i := len(b) - 1
	for i > 0 {
		out[i] = b[i] ^ b[i-1]
		i--
	}
	out[0] = b[0] ^ 0x6B
	return out
}

func Encode(b []byte) []byte {
	if len(b) < 1 {
		return []byte{}
	}
	out := make([]byte, len(b))
	out[0] = b[0] ^ 0x6B
	i := 1
	for i < len(b) {
		out[i] = b[i] ^ out[i-1]
		i++
	}
	return out
}

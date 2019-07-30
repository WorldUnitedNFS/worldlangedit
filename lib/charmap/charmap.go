// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package charmap

import (
	"encoding/binary"
	"fmt"
)

type Charmap struct {
	chunk []byte
}

func FromChunk(chunk []byte) *Charmap {
	return &Charmap{
		chunk: chunk,
	}
}

func (cm *Charmap) DecodeBytes(b []byte) string {
	out := make([]rune, 0)
	i := 0
	for {
		if i == len(b) {
			break
		}
		if i > len(b) {
			panic("i is out of bounds")
		}
		char := b[i]
		if char < 128 {
			out = append(out, rune(char))
			i++
		} else {
			offset := 12 + int(char)*2
			newChar := binary.LittleEndian.Uint16(cm.chunk[offset : offset+2])
			if newChar < 128 {
				addCount := 128 * int(newChar)
				char = b[i+1]
				offset := 12 + int(char)*2 + addCount
				newChar := binary.LittleEndian.Uint16(cm.chunk[offset : offset+2])
				out = append(out, rune(newChar))
				i += 2
			} else {
				out = append(out, rune(newChar))
				i++
			}
		}
	}
	return string(out)
}

func (cm *Charmap) findHighCharOffset(char rune) int {
	offset := 0
	for {
		byteOffset := 12 + offset*2
		if byteOffset == len(cm.chunk) {
			panic(fmt.Sprintf("Cannot encode character %x (%c)", char, char))
		}
		charThere := binary.LittleEndian.Uint16(cm.chunk[byteOffset : byteOffset+2])
		if rune(charThere) == char {
			break
		}
		offset++
	}
	return offset
}

func (cm *Charmap) EncodeString(str string) []byte {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Printf("Failed to encode string %s\n", str)
			panic(err)
		}
	}()
	out := make([]byte, 0)
	for _, char := range str {
		if char < 128 {
			out = append(out, byte(char))
		} else {
			offset := cm.findHighCharOffset(char)
			if offset > 255 {
				headCharCnt := offset / 128
				offset -= (headCharCnt - 1) * 128
				headChar := cm.findHighCharOffset(rune(headCharCnt))
				out = append(out, byte(headChar))
			}
			if offset > 255 {
				panic("offset overflows")
			}
			out = append(out, byte(offset))
		}
	}
	return out
}

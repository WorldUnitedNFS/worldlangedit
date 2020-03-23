// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package charmap

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Charmap struct {
	NumEntries int32
	EntryTable [0xC00]uint16
}

func FromChunk(chunk []byte) *Charmap {
	buf := bytes.NewBuffer(chunk)
	cm := &Charmap{}
	err := binary.Read(buf, binary.LittleEndian, cm)
	if err != nil {
		panic(err)
	}
	return cm
}

func (cm *Charmap) DecodeBytes(b []byte) string {
	runes := make([]rune, 0)

	for i := 0; i < len(b); {
		curByte := rune(b[i])
		i++

		if curByte >= 0x80 {
			histEntry := rune(cm.EntryTable[curByte])

			if histEntry >= 0x80 {
				curByte = histEntry
				//fmt.Printf("encountered: %c\n", curByte)
			} else if histEntry != 0 {
				nextByte := b[i]
				i++
				if nextByte >= 0x80 {
					curByte = rune(cm.EntryTable[128*histEntry-128+rune(nextByte)])
					//fmt.Printf("encountered: %c\n", curByte)
				}
			} else {
				panic("Could not decode string")
			}
		}

		runes = append(runes, curByte)
	}

	return string(runes)
}

func (cm *Charmap) EncodeString(str string) []byte {
	out := make([]byte, 0)

	for _, c := range str {
		if c >= 0xFF80 {
			panic("what even IS this?")
		}

		savedChar := c

		if c >= 0x80 {
			curIndex := int32(128)
			maxIndex := cm.NumEntries

			if cm.NumEntries > 128 {
				for curIndex < maxIndex {
					if rune(cm.EntryTable[curIndex]) == c {
						break
					}
					curIndex++
				}
			}

			if curIndex >= 256 {
				if curIndex != maxIndex {
					c = 128
					searchIndex := 128
					update := true

					for rune(cm.EntryTable[searchIndex]) != curIndex>>7 {
						c++
						if c >= 256 {
							update = false
							break
						}
						searchIndex++
					}

					if update {
						out = append(out, byte(c))
						out = append(out, byte(curIndex%128-128))
					}
				}

				notFound := c == 256 || curIndex == maxIndex
				c = savedChar
				if notFound {
					panic(fmt.Sprintf("could not encode character %c (%d)! string: %s", c, c, str))
				}
			} else {
				out = append(out, byte(curIndex))
			}
		} else {
			out = append(out, byte(c))
		}
	}

	return out
}

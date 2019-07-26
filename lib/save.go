// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lib

import (
	"encoding/binary"
	"sort"
	"strings"

	"github.com/redbluescreen/worldlangedit/lib/xor"
	"golang.org/x/text/encoding/charmap"
)

func SaveFile(file *LangFile, lFile *LangFile, doXor bool) []byte {
	encoder := charmap.ISO8859_1.NewEncoder()

	type hEntry struct {
		Hash   uint32
		String []byte
		Label  string
		Offset uint32
	}
	hEntries := make([]*hEntry, len(file.Entries))
	for i, e := range file.Entries {
		b, err := encoder.Bytes([]byte(e.String))
		if err != nil {
			panic(err)
		}
		var label string
		for _, le := range lFile.Entries {
			if le.Hash == e.Hash {
				label = le.String
			}
		}
		hEntries[i] = &hEntry{
			Hash:   e.Hash,
			String: b,
			Label:  label,
		}
	}

	stringsLen := 0
	for _, e := range hEntries {
		stringsLen += len(e.String) + 1
	}
	langLen := 36 + len(hEntries)*8 + stringsLen
	langLen += 4 - (langLen % 4)
	paddingLen := 16 - (langLen % 16)
	if paddingLen-8 < 0 {
		paddingLen += 16
	}
	totalLen := langLen + paddingLen + len(file.EndData)

	data := make([]byte, totalLen)
	binary.LittleEndian.PutUint32(data[0:4], 0x39000)
	binary.LittleEndian.PutUint32(data[4:8], uint32(langLen-8))
	binary.LittleEndian.PutUint32(data[8:12], uint32(len(hEntries)))
	binary.LittleEndian.PutUint32(data[12:16], 0x1C)
	binary.LittleEndian.PutUint32(data[16:20], uint32(len(hEntries)*8+28))
	copy(data[20:36], []byte{'G', 'l', 'o', 'b', 'a', 'l'})

	sort.SliceStable(hEntries, func(i, j int) bool {
		l1 := strings.ToLower(hEntries[i].Label)
		l2 := strings.ToLower(hEntries[j].Label)
		// return l1 < l2
		// l1 := hEntries[i].Label
		// l2 := hEntries[j].Label
		// if len(l1) < len(l2) {
		// 	return true
		// }
		// if len(l1) > len(l2) {
		// 	return false
		// }
		for i = 0; i < len(l1) && i < len(l2); i++ {
			c1 := l1[i]
			c2 := l2[i]
			// if c1 == '_' && c2 >= 48 && c2 != '_' {
			// 	return true
			// }
			// if c1 >= 48 && c1 != '_' && c2 == '_' {
			// 	return false
			// }
			if c1 == '.' {
				c1 = '-'
			} else if c1 == '_' {
				c1 = '.'
			} else if c1 == '-' {
				c1 = '_'
			}

			if c2 == '.' {
				c2 = '-'
			} else if c2 == '_' {
				c2 = '.'
			} else if c2 == '-' {
				c2 = '_'
			}

			if c1 < c2 {
				return true
			}
			if c1 > c2 {
				return false
			}
		}
		if len(l1) < len(l2) {
			return true
		}
		return false
	})
	offset := 36 + len(hEntries)*8
	inOffset := 0
	for _, e := range hEntries {
		e.Offset = uint32(inOffset)
		copy(data[offset:], e.String)
		data[offset+len(e.String)] = 0
		offset += len(e.String) + 1
		inOffset += len(e.String) + 1
	}

	sort.SliceStable(hEntries, func(i, j int) bool {
		return hEntries[i].Hash < hEntries[j].Hash
	})
	offset = 36
	for _, e := range hEntries {
		binary.LittleEndian.PutUint32(data[offset:offset+4], e.Hash)
		binary.LittleEndian.PutUint32(data[offset+4:offset+8], e.Offset)
		offset += 8
	}

	binary.LittleEndian.PutUint32(data[langLen+4:langLen+8], uint32(paddingLen-8))
	copy(data[langLen+paddingLen:], file.EndData)

	if doXor {
		return xor.Encode(data)
	}

	return data
}

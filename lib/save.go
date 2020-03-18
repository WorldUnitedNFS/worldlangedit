// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lib

import (
	"encoding/binary"
	"github.com/WorldUnitedNFS/worldlangedit/lib/charmap"
	"github.com/WorldUnitedNFS/worldlangedit/lib/xor"
	"sort"
	"strings"
)

func SaveFile(file *LangFile, lFile *LangFile, doXor bool) []byte {
	// char map debug
	cm := BuildCharMap(file)

	type hEntry struct {
		Hash       uint32
		String     []byte
		Label      string
		Offset     uint32
		OrigString string
	}

	hEntries := make([]*hEntry, len(file.Entries))
	for i, e := range file.Entries {
		var label string
		for _, le := range lFile.Entries {
			if le.Hash == e.Hash {
				label = le.String
				break
			}
		}
		//fmt.Printf("encoded %s in %d bytes\n", e.String, len(b))
		hEntries[i] = &hEntry{
			Hash:       e.Hash,
			Label:      label,
			OrigString: e.String,
		}
	}
	sort.SliceStable(hEntries, func(i, j int) bool {
		l1 := strings.ToLower(hEntries[i].Label)
		l2 := strings.ToLower(hEntries[j].Label)

		for i = 0; i < len(l1) && i < len(l2); i++ {
			c1 := l1[i]
			c2 := l2[i]

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

	for i, e := range file.Entries {
		b := cm.EncodeString(e.String)
		var label string
		for _, le := range lFile.Entries {
			if le.Hash == e.Hash {
				label = le.String
				break
			}
		}
		//fmt.Printf("encoded %s in %d bytes\n", e.String, len(b))
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
	totalLen := langLen + paddingLen + /*endData*/ 8 + 4 + (0xC00 * 2)

	data := make([]byte, totalLen)
	binary.LittleEndian.PutUint32(data[0:4], 0x39000)
	binary.LittleEndian.PutUint32(data[4:8], uint32(langLen-8))
	binary.LittleEndian.PutUint32(data[8:12], uint32(len(hEntries)))
	binary.LittleEndian.PutUint32(data[12:16], 0x1C)
	binary.LittleEndian.PutUint32(data[16:20], uint32(len(hEntries)*8+28))
	copy(data[20:36], []byte{'G', 'l', 'o', 'b', 'a', 'l'})

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
	//copy(data[langLen+paddingLen:], )
	binary.LittleEndian.PutUint32(data[langLen+paddingLen:], 0x39001)
	binary.LittleEndian.PutUint32(data[langLen+paddingLen+4:], 0x1804)
	binary.LittleEndian.PutUint32(data[langLen+paddingLen+8:], uint32(cm.NumEntries))
	for i := 0; i < 0xC00; i++ {
		binary.LittleEndian.PutUint16(data[langLen+paddingLen+12+2*i:], cm.EntryTable[i])
	}

	if doXor {
		return xor.Encode(data)
	}

	return data
}

func BuildCharMap(lf *LangFile) charmap.Charmap {
	// build our own char map
	newCharMap := charmap.Charmap{
		NumEntries: 0,
		EntryTable: [3072]uint16{},
	}

	numEntries := int32(0x80)      // 0x00-0x7F get reserved spaces
	charSet := make(map[rune]bool) // keeps track of characters we've encountered

	for _, entry := range lf.Entries {
		for _, c := range entry.String {
			if c >= 0x80 {
				charSet[c] = true
			}
		}
	}

	chars := make([]rune, 0)

	for r := range charSet {
		chars = append(chars, r)
	}

	sort.SliceStable(chars, func(i, j int) bool {
		return chars[j] < chars[i]
	})

	numEntries += int32(len(chars))

	jumpEntries := make([]uint16, 0)
	maxJumpEntry := numEntries >> 7

	if maxJumpEntry >= 2 {
		jumpEntries = append(jumpEntries, uint16(maxJumpEntry))
		numEntries++
	}

	// Determine jump entries
	for {
		newMaxJumpEntry := numEntries >> 7

		if newMaxJumpEntry > maxJumpEntry {
			numEntries++
			maxJumpEntry = newMaxJumpEntry
			jumpEntries = append(jumpEntries, uint16(maxJumpEntry))
		} else {
			break
		}
	}

	mapIndex := 0x80

	for i := len(jumpEntries) - 1; i >= 0; i-- {
		newCharMap.EntryTable[mapIndex] = jumpEntries[i]
		mapIndex++
	}

	for _, r := range chars {
		newCharMap.EntryTable[mapIndex] = uint16(r)
		mapIndex++
	}

	newCharMap.NumEntries = numEntries

	return newCharMap
}

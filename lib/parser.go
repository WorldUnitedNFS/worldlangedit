// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lib

import (
	"encoding/binary"

	"github.com/redbluescreen/worldlangedit/lib/charmap"
	"github.com/redbluescreen/worldlangedit/lib/xor"
)

func ztString(b []byte) []byte {
	var sb []byte
	for i, c := range b {
		if c == 0 {
			sb = b[:i]
			break
		}
	}
	return sb
}

func ParseFile(data []byte) *LangFile {
	if IsFileEncoded(data) {
		data = xor.Decode(data)
	}

	chunkLen := binary.LittleEndian.Uint32(data[4:8])
	entryCount := binary.LittleEndian.Uint32(data[8:12])
	entries := make([]LangFileEntry, int(entryCount))

	stringsStart := binary.LittleEndian.Uint32(data[16:20])

	offset := int(chunkLen) + 8
	// Chunk type 0 -> padding
	if binary.LittleEndian.Uint32(data[offset:offset+4]) == 0 {
		offset += int(binary.LittleEndian.Uint32(data[offset+4:offset+8])) + 8
	}
	endData := data[offset:]
	chm := charmap.FromChunk(endData)

	offset = 36
	entryIdx := 0
	for {
		hash := binary.LittleEndian.Uint32(data[offset : offset+4])
		location := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		strBytes := ztString(data[stringsStart+8+location:])
		entries[entryIdx] = LangFileEntry{
			Hash:          hash,
			String:        chm.DecodeBytes(strBytes),
			OriginalBytes: strBytes,
			Offset:        location,
		}
		entryIdx++
		offset += 8
		if offset == int(stringsStart+8) {
			break
		}
	}

	return &LangFile{
		Entries: entries,
		EndData: endData,
	}
}

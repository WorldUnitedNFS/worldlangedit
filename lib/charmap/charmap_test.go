// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package charmap_test

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"

	"github.com/redbluescreen/worldlangedit/lib"
	"github.com/redbluescreen/worldlangedit/lib/charmap"
)

func TestDecodeEncode(t *testing.T) {
	languages := []string{
		"Chinese_Simp",
		"Chinese_Trad",
		"English",
		"French",
		"German",
		"Polish",
		"Portuguese",
		"Russian",
		"Spanish",
		"Thai",
		"Turkish",
	}
	var failureFile *os.File
	defer func() {
		if failureFile != nil {
			failureFile.Close()
		}
	}()
	for _, language := range languages {
		data, err := ioutil.ReadFile("../../testdata/" + language + "_Global.bin")
		if err != nil {
			t.Skip("Failed to read file for language " + language)
			return
		}
		f := lib.ParseFile(data)
		chm := charmap.FromChunk(f.EndData)
		correct := 0
		for _, entry := range f.Entries {
			enc := chm.EncodeString(entry.String)
			if bytes.Equal(enc, entry.OriginalBytes) {
				correct++
			} else {
				if failureFile == nil {
					failureFile, err = os.Create("encode_failure.log")
					if err != nil {
						panic(err)
					}
				}
				failureFile.WriteString("----------------------------------------\n")
				failureFile.WriteString("Original bytes:\n")
				failureFile.WriteString(hex.Dump(entry.OriginalBytes))
				failureFile.WriteString("Decoded string: " + entry.String + "\n")
				failureFile.WriteString("Encoded bytes:\n")
				failureFile.WriteString(hex.Dump(enc))
			}
		}
		t.Logf("%s: %v/%v (%.1f%%)", language, correct, len(f.Entries), float64(correct)/float64(len(f.Entries))*100)
		if correct != len(f.Entries) {
			t.Fail()
		}
	}
}

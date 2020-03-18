package charmap_test

import (
	"github.com/WorldUnitedNFS/worldlangedit/lib"
	"io/ioutil"
	"testing"
)

func TestDecode(t *testing.T) {
	bytes, err := ioutil.ReadFile("../../testdata/Chinese_Simp_Global.bin")

	if err != nil {
		t.Errorf("Error occurred while reading file: %v", err)
		t.FailNow()
	}

	lf := lib.ParseFile(bytes)
	entry := lf.Entries[0]

	decodedText := lf.CharMap.DecodeBytes(entry.OriginalBytes)

	if decodedText != "激活工作人员" {
		t.Errorf("Failed to decode string. Expected '激活工作人员', got %s", decodedText)
		t.FailNow()
	}
}

func TestMakingOurOwn(t *testing.T) {
	// Load French file
	langBytes, err := ioutil.ReadFile("../../testdata/French_Global.bin")

	if err != nil {
		t.Errorf("Error occurred while reading file: %v", err)
		t.FailNow()
	}
	labelsBytes, err := ioutil.ReadFile("../../testdata/Labels_Global.bin")

	if err != nil {
		t.Errorf("Error occurred while reading file: %v", err)
		t.FailNow()
	}

	langFile := lib.ParseFile(langBytes)
	labelsFile := lib.ParseFile(labelsBytes)

	t.Logf("Loaded language file. Histogram has %d entries", langFile.CharMap.NumEntries)

	for i, c := range langFile.CharMap.EntryTable {
		if c >= 0x80 {
			t.Logf("\tHistogram entry [%d]: %c (0x%04x) (0b%16b)", i, c, c, c)
		}
	}

	newMap := lib.BuildCharMap(langFile)

	t.Logf("New histogram has %d entries", newMap.NumEntries)

	for _, e := range langFile.Entries {
		encodedBytes := newMap.EncodeString(e.String)
		decodedText := newMap.DecodeBytes(encodedBytes)

		if decodedText != e.String {
			t.Errorf("String 0x%08x (%d) was corrupted. Original: %s / New: %s", e.Hash, e.Hash, e.String, decodedText)
			t.FailNow()
		}

		t.Logf("String 0x%08x (%d) encoded/decoded successfully", e.Hash, e.Hash)
	}

	// Save file and test
	t.Logf("Testing generated data...")

	genData := lib.SaveFile(langFile, labelsFile, true)
	loadedFromGen := lib.ParseFile(genData)

	for _, e := range langFile.Entries {
		e2 := loadedFromGen.FindEntryByHash(e.Hash)

		if e2.String != e.String {
			t.Errorf("String 0x%08x (%d) was corrupted. Original: %s / New: %s", e.Hash, e.Hash, e.String, e2.String)
			t.FailNow()
		}
	}
}

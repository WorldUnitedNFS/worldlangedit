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

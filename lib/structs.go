// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lib

import "github.com/WorldUnitedNFS/worldlangedit/lib/charmap"

type LangFile struct {
	Entries []LangFileEntry
	CharMap *charmap.Charmap
}

type LangFileEntry struct {
	Hash          uint32
	String        string
	Offset        uint32
	OriginalBytes []byte
}

func (lf *LangFile) FindEntryByHash(hash uint32) *LangFileEntry {
	for _, e := range lf.Entries {
		if e.Hash == hash {
			return &e
		}
	}

	return nil
}

func (lf *LangFile) FindEntryByName(name string) *LangFileEntry {
	return lf.FindEntryByHash(BinHash(name))
}

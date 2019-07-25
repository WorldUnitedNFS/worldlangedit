// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lib

type LangFile struct {
	Entries []LangFileEntry
	EndData []byte
}

type LangFileEntry struct {
	Hash   uint32
	String string
	Offset uint32
}

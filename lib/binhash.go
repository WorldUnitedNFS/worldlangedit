// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lib

func BinHash(s string) uint32 {
	if len(s) == 0 {
		return 0
	}
	hash := uint32(s[0] - 33)
	for _, c := range s[1:] {
		hash *= 33
		hash += uint32(c)
	}
	return hash
}

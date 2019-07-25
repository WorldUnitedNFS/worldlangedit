// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lib

import "bytes"

func IsFileEncoded(b []byte) bool {
	return !bytes.Equal(
		b[20:26],
		[]byte{'G', 'l', 'o', 'b', 'a', 'l'},
	)
}

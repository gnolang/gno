// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha3

import (
	isha3 "internal/crypto/sha3"
)


func Sum256(data []byte) []byte {
	return isha3.Sum256(data)
}

func Sum512(data []byte) []byte {
	return isha3.Sum512(data)
}
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha3

import (
	isha3 "internal/crypto/sha3"
)

// Sum256 returns the SHA3-256 checksum of the data.
func Sum256(data []byte) [32]byte {
	return isha3.Sum256(data)
}

// Sum512 returns the SHA3-512 checksum of the data.
func Sum512(data []byte) [64]byte {
	return isha3.Sum512(data)
}
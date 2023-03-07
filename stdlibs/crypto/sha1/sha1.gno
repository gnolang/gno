// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sha1 implements the SHA-1 hash algorithm as defined in RFC 3174.
//
// SHA-1 is cryptographically broken and should not be used for secure
// applications.

package sha1

import (
	isha1 "internal/crypto/sha1" // XXX
)

// Sum returns the SHA-1 checksum of the data.
func Sum(data string) []byte {
	return isha1.Sum(data)
}
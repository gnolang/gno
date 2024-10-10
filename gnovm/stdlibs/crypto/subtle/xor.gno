// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// GNO modifications:
// Modified to remove unsafe code for deterministic behavior for for gnovm
// This gives the same result as the original code, but is not as efficient.
// original behavior can be found here:
// https://cs.opensource.google/go/go/+/master:src/crypto/subtle/xor.go

package subtle

// XORBytes sets dst[i] = x[i] ^ y[i] for all i < n = min(len(x), len(y)),
// returning n, the number of bytes written to dst.
// If dst does not have length at least n,
// XORBytes panics without writing anything to dst.
func XORBytes(dst, x, y []byte) int {
	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if n == 0 {
		return 0
	}
	if n > len(dst) {
		panic("subtle.XORBytes: dst too short")
	}

	for i := 0; i < n; i++ {
		dst[i] = x[i] ^ y[i]
	}
	return n
}

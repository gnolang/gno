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
	/* XXX: removed due to min function dependency on generics.
	n := min(len(x), len(y))
	*/

	// XXX: reverted to go1.22 behavior
	// https://cs.opensource.google/go/go/+/refs/tags/go1.22.0:src/crypto/subtle/xor.go;l=12
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

	/* XXX: removed due to dependencies on runtime and unsafe packages
	        https://cs.opensource.google/go/go/+/refs/tags/go1.22.0:src/crypto/subtle/xor_generic.go
	xorBytes(&dst[0], &x[0], &y[0], n) // arch-specific
	*/

	// XXX: replaced with pure Go implementation.
	for i := 0; i < n; i++ {
		dst[i] = x[i] ^ y[i]
	}
	return n
}

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtle

// XORBytes sets dst[i] = x[i] ^ y[i] for all i < n = min(len(x), len(y)),
// returning n, the number of bytes written to dst.
//
// If dst does not have length at least n,
// XORBytes panics without writing anything to dst.
//
// dst and x or y may overlap exactly or not at all,
// otherwise XORBytes may panic.
func XORBytes(dst, x, y []byte) int {
	n := min(len(x), len(y))
	if n == 0 {
		return 0
	}
	if n > len(dst) {
		panic("subtle.XORBytes: dst too short")
	}
	if inexactOverlap(dst[:n], x[:n]) || inexactOverlap(dst[:n], y[:n]) {
		panic("subtle.XORBytes: invalid overlap")
	}

	xorBytes(dst, x, y, n)
	return n
}

// xorBytes XORs bytes from x and y into dst.
// All slices must have length >= n.
func xorBytes(dst, x, y []byte, n int) {
	for i := 0; i < n; i++ {
		dst[i] = x[i] ^ y[i]
	}
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// XXX this section is an adaptation of crypto/internal/fips140/alias without using 'unsafe'
// anyOverlap reports whether x and y share memory at any (not necessarily
// corresponding) index. The memory beyond the slice length is ignored.
func anyOverlap(x, y []byte) bool {
	if len(x) == 0 || len(y) == 0 {
		return false
	}

	for i := range x {
		for j := range y {
			if &x[i] == &y[j] {
				return true
			}
		}
	}
	return false
}

// InexactOverlap reports whether x and y share memory at any non-corresponding
// index. The memory beyond the slice length is ignored. Note that x and y can
// have different lengths and still not have any inexact overlap.
//
// InexactOverlap can be used to implement the requirements of the crypto/cipher
// AEAD, Block, BlockMode and Stream interfaces.
func inexactOverlap(x, y []byte) bool {
	if len(x) == 0 || len(y) == 0 || &x[0] == &y[0] {
		return false
	}
	return anyOverlap(x, y)
}

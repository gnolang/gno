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
func XORBytes(x, y []byte) []byte {
	n := min(len(x), len(y))
	buf := make([]byte, n)
	xorBytes(buf, x, y, n)
	return buf
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

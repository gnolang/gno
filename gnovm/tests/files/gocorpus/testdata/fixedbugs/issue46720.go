// compile

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f() {
	nonce := make([]byte, 24)
	g((*[24]byte)(nonce))
}

//go:noinline
func g(*[24]byte) {}

// GnoPreprocessError:
// line 11: cannot convert nonce<VPBlock(1,0)> (of type []uint8) to type *[24]uint8

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects code gc + go/types accept)

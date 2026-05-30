// compile

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Caused a gofrontend crash.

package p

func F(b []byte, i int) {
	*(*[1]byte)(b[i*2:]) = [1]byte{}
}

// KnownIssue:
// line 12: cannot convert b<VPBlock(1,0)>[i<VPBlock(1,1)> * (const (2 int)):] (of type []uint8) to type *[1]uint8

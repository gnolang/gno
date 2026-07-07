// compile

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 59169: caused gofrontend crash.

package p

func F(p *[]byte) {
	*(*[1]byte)(*p) = *(*[1]byte)((*p)[1:])
}

// GnoPreprocessError:
// line 12: cannot convert *(p<VPBlock(1,0)>)[(const (1 int)):] (of type []uint8) to type *[1]uint8

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects code gc + go/types accept)

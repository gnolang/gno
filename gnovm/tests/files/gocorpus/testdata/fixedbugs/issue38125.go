// compile

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// gccgo mishandled embedded methods of type aliases.

package p

type I int

func (I) M() {}

type T = struct {
	I
}

func F() {
	_ = T.M
	_ = struct { I }.M
}

// GnoError:
// line 20: invalid method expression T.M (type struct{I gno.land/p/filetest/p.I} has no method M)

// Verdict: KnownIssue compile: #5787 method expression on a promoted method is rejected at preprocess

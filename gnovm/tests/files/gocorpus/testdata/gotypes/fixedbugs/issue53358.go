// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type A struct{}

func (*A) m() int { return 0 }

var _ = A.m /* ERROR "invalid method expression A.m (needs pointer receiver (*A).m)" */ ()
var _ = (*A).m(nil)

type B struct{ A }

var _ = B.m // ERROR "invalid method expression B.m (needs pointer receiver (*B).m)"
var _ = (*B).m

var _ = struct{ A }.m // ERROR "invalid method expression struct{A}.m (needs pointer receiver (*struct{A}).m)"

// GnoError:
// line 11: invalid method expression A.m (needs pointer receiver (*A).m)
// line 16: invalid method expression B.m (needs pointer receiver (*B).m)
// line 19: struct {  typeval{gno.land/p/filetest/p.A} }.m undefined (type struct{A gno.land/p/filetest/p.A} has no method m)

// GnoWordingDiffers:
// line 19: gc expects: invalid method expression struct{A}.m (needs pointer receiver (*struct{A}).m); gno: struct {  typeval{gno.land/p/filetest/p.A} }.m undefined (type struct{A gno.land/p/filetest/p.A} has no method m)

// Verdict: Clean errorcheck

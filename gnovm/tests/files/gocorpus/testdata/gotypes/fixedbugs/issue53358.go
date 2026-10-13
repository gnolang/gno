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
// line 11: wrong argument count in call to typeval{gno.land/p/filetest/p.A}.m; want 1 got 0
// line 16: unknown *DeclaredType method named m
// line 17: unknown *DeclaredType method named m
// line 19: unexpected selector expression type value struct{A gno.land/p/filetest/p.A}

// GoTypeCheckError:
// line 11
// line 16
// line 19

// GnoOverStrictError:
// line 17: unknown *DeclaredType method named m

// KnownIssue:
// errorcheck: 1 over-strict line(s), 0 uncaught marker(s)

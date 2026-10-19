// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T struct{}
type A = T
type B = T

func (T) m() {}
func (T) m() {} // ERROR "already declared|redefinition"
func (A) m() {} // ERROR "already declared|redefinition"
func (A) m() {} // ERROR "already declared|redefinition"
func (B) m() {} // ERROR "already declared|redefinition"
func (B) m() {} // ERROR "already declared|redefinition"

func (*T) m() {} // ERROR "already declared|redefinition"
func (*A) m() {} // ERROR "already declared|redefinition"
func (*B) m() {} // ERROR "already declared|redefinition"

// GnoError:
// line 14: T.m redeclared in this block; other declaration at gno.land/p/filetest/p/issue18655.go:14:1-16
// line 15: redeclaration of method T.m
// line 16: A.m redeclared in this block; other declaration at gno.land/p/filetest/p/issue18655.go:16:1-16
// line 17: redeclaration of method T.m
// line 18: B.m redeclared in this block; other declaration at gno.land/p/filetest/p/issue18655.go:18:1-16
// line 20: T.m redeclared in this block; other declaration at gno.land/p/filetest/p/issue18655.go:14:1-16
// line 21: A.m redeclared in this block; other declaration at gno.land/p/filetest/p/issue18655.go:16:1-16
// line 22: B.m redeclared in this block; other declaration at gno.land/p/filetest/p/issue18655.go:18:1-16

// Verdict: KnownIssue errorcheck: #6344 on-chain path accepts the file

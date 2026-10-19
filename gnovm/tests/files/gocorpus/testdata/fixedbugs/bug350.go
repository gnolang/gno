// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T int

func (T) m() {} // GCCGO_ERROR "previous"
func (T) m() {} // ERROR "T\.m already declared|redefinition"

func (*T) p() {} // GCCGO_ERROR "previous"
func (*T) p() {} // ERROR "T\.p already declared|redefinition"

// GnoError:
// line 12: T.m redeclared in this block; other declaration at main/bug350.go:11:1-16
// line 15: T.p redeclared in this block; other declaration at main/bug350.go:14:1-17

// GnoWordingDiffers:
// line 12: gc expects: T\.m already declared|redefinition; gno: T.m redeclared in this block; other declaration at main/bug350.go:11:1-16
// line 15: gc expects: T\.p already declared|redefinition; gno: T.p redeclared in this block; other declaration at main/bug350.go:14:1-17

// Verdict: KnownIssue errorcheck: #6344 on-chain path accepts the file

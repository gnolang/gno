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
// line 12: redeclaration of method T.m
// line 15: redeclaration of method T.p

// GoTypeCheckError:
// line 12: method T.m already declared at main/bug350.go:11:10
// line 15: method T.p already declared at main/bug350.go:14:11

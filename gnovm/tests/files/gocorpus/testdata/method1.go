// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that method redeclarations are caught by the compiler.
// Does not compile.

package main

type T struct{}

func (t *T) M(int, string)  // GCCGO_ERROR "previous"
func (t *T) M(int, float64) {} // ERROR "already declared|redefinition"

func (t T) H()  // GCCGO_ERROR "previous"
func (t *T) H() {} // ERROR "already declared|redefinition"

func f(int, string)  // GCCGO_ERROR "previous"
func f(int, float64) {} // ERROR "redeclared|redefinition"

func g(a int, b string) // GCCGO_ERROR "previous"
func g(a int, c string) // ERROR "redeclared|redefinition"

// GnoError:
// line 15: redeclaration of method T.M
// line 18: redeclaration of method T.H
// line 20: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 21: f redeclared in this block
// 	previous declaration at method1.go:20:6 (and 1 more errors)
// line 23: function g does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 24: g redeclared in this block
// 	previous declaration at method1.go:23:6

// GoTypeCheckError:
// line 15: method T.M already declared at main/method1.go:14:13
// line 18: method T.H already declared at main/method1.go:17:12

// GnoOverStrictError:
// line 20: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 23: function g does not have a body but is not natively defined (did you build after pulling from the repository?)

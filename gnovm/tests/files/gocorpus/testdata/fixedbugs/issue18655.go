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
// line 14: redeclaration of method T.m
// line 15: redeclaration of method T.m
// line 16: redeclaration of method T.m
// line 17: redeclaration of method T.m
// line 18: redeclaration of method T.m
// line 20: redeclaration of method T.m
// line 21: redeclaration of method T.m
// line 22: redeclaration of method T.m

// Verdict: Clean errorcheck

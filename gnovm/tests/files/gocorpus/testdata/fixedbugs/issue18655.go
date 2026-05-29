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

// GoTypeCheckError:
// line 14: method T.m already declared at gno.land/p/filetest/p/issue18655.go:14:10
// line 15: method T.m already declared at gno.land/p/filetest/p/issue18655.go:14:10
// line 16: method T.m already declared at gno.land/p/filetest/p/issue18655.go:14:10
// line 17: method T.m already declared at gno.land/p/filetest/p/issue18655.go:14:10
// line 18: method T.m already declared at gno.land/p/filetest/p/issue18655.go:14:10
// line 20: method T.m already declared at gno.land/p/filetest/p/issue18655.go:14:10
// line 21: method T.m already declared at gno.land/p/filetest/p/issue18655.go:14:10
// line 22: method T.m already declared at gno.land/p/filetest/p/issue18655.go:14:10

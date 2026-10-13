// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

type I interface {
	Foo()
}

type T1 struct{}

func (T1) foo() {}

type T2 struct{}

func (T2) foo() string { return "" }

func _() {
	var i I
	_ = i /* ERROR "impossible type assertion: i.(T1)\n\tT1 does not implement I (missing method Foo)\n\t\thave foo()\n\t\twant Foo()" */ .(T1)
	_ = i /* ERROR "impossible type assertion: i.(T2)\n\tT2 does not implement I (missing method Foo)\n\t\thave foo() string\n\t\twant Foo()" */ .(T2)
}

// GnoError:
// line 21: impossible type assertion: gno.land/p/filetest/pkg.T1 does not implement gno.land/p/filetest/pkg.I (missing method Foo)
// line 22: impossible type assertion: gno.land/p/filetest/pkg.T2 does not implement gno.land/p/filetest/pkg.I (missing method Foo)

// GoTypeCheckError:
// line 21
// line 22

// Clean: errorcheck

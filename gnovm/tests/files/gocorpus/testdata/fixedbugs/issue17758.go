// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func foo() {
	_ = func() {}
}

func foo() { // ERROR "foo redeclared in this block|redefinition of .*foo.*"
	_ = func() {}
}

func main() {}

// GnoError:
// line 13: foo redeclared in this block
// 	previous declaration at issue17758.go:9:6

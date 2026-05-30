// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check that we don't print duplicate errors for string ->
// array-literal conversion

package main

func main() {
	_ = []byte{"foo"}   // ERROR "cannot use|incompatible type|cannot convert"
	_ = []int{"foo"}    // ERROR "cannot use|incompatible type|cannot convert"
	_ = []rune{"foo"}   // ERROR "cannot use|incompatible type|cannot convert"
	_ = []string{"foo"} // OK
}

// GnoError:
// line 13: cannot use untyped string as Uint8Kind
// line 14: cannot use untyped string as IntKind
// line 15: cannot use untyped string as Int32Kind

// GoTypeCheckError:
// line 13: cannot use "foo" (untyped string constant) as byte value in array or slice literal
// line 14: cannot use "foo" (untyped string constant) as int value in array or slice literal
// line 15: cannot use "foo" (untyped string constant) as rune value in array or slice literal

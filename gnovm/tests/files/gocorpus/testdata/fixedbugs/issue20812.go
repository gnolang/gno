// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f() {
	_ = int("1")      // ERROR "cannot convert|invalid type conversion"
	_ = bool(0)       // ERROR "cannot convert|invalid type conversion"
	_ = bool("false") // ERROR "cannot convert|invalid type conversion"
	_ = int(false)    // ERROR "cannot convert|invalid type conversion"
	_ = string(true)  // ERROR "cannot convert|invalid type conversion"
}

// GnoError:
// line 10: cannot convert StringKind to IntKind
// line 11: cannot convert IntKind to BoolKind
// line 12: cannot convert StringKind to BoolKind
// line 13: cannot convert BoolKind to IntKind
// line 14: cannot convert BoolKind to StringKind

// GoTypeCheckError:
// line 10: cannot convert "1" (untyped string constant) to type int
// line 11: cannot convert 0 (untyped int constant) to type bool
// line 12: cannot convert "false" (untyped string constant) to type bool
// line 13: cannot convert false (untyped bool constant) to type int
// line 14: cannot convert true (untyped bool constant) to type string

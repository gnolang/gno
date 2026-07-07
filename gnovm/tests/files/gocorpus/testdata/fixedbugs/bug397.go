// errorcheck -d=panic

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Issue 2623
var m = map[string]int{
	"abc": 1,
	1:     2, // ERROR "cannot use 1.*as type string in map key|incompatible type|cannot convert|cannot use"
}

// GnoError:
// line 10: 2: cannot use untyped Bigint as StringKind
// line 11: expected declaration, found "abc"
// line 12: expected declaration, found 1
// line 13: expected declaration, found '}'

// GoTypeCheckError:
// line 12: cannot use 1 (untyped int constant) as string value in map literal

// GnoOverStrictError:
// line 10: 2: cannot use untyped Bigint as StringKind
// line 11: expected declaration, found "abc"
// line 13: expected declaration, found '}'

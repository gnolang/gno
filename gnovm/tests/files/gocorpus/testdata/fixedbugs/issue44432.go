// errorcheck -d=panic

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var m = map[string]int{
	"a": 1,
	1:   1, // ERROR "cannot use 1.*as.*string.*in map"
	2:   2, // ERROR "cannot use 2.*as.*string.*in map"
}

// GnoOverStrictError:
// line 9: 2: cannot use untyped Bigint as StringKind

// GoTypeCheckError:
// line 11: cannot use 1 (untyped int constant) as string value in map literal
// line 12: cannot use 2 (untyped int constant) as string value in map literal

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)

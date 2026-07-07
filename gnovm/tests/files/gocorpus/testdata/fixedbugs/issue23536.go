// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test case where a slice of a user-defined byte type (not uint8 or byte) is
// converted to a string.  Same for slice of runes.

package main

type MyByte byte

type MyRune rune

func main() {
	var y []MyByte
	_ = string(y)

	var z []MyRune
	_ = string(z)
}


// Fixed: master PR #5780 (5d7ec8679); verified clean, broken at parent; re-golden after rebase.

// GnoOutput:

// GnoError:
// interface conversion: gnolang.Value is nil, not *gnolang.SliceValue

// GoOutput:

// KnownIssue:
// Converting a nil slice of a named byte/rune type to string crashed the
// VM: the nil slice's Value was asserted to *SliceValue without a nil check.

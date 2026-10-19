// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Literals that happen to resolve to named constants
// may be used as label names (see issue 13684). Make
// sure that other literals don't crash the compiler.

package main

const labelname = 1

func main() {
	goto labelname
labelname:
}

func f() {
	var x int
	switch x {
	case 1:
		2:	// ERROR "unexpected :|expected .*;.* or .*}.* or newline|value computed is not used"
	case 2:
	}

	switch x {
	case 1:
		2: ;	// ERROR "unexpected :|expected .*;.* or .*}.* or newline|value computed is not used"
	case 2:
	}

	var y string
	switch y {
	case "foo":
		"bar":	// ERROR "unexpected :|expected .*;.* or .*}.* or newline|value computed is not used"
	case "bar":
	}

	switch y {
	case "foo":
		"bar": ;	// ERROR "unexpected :|expected .*;.* or .*}.* or newline|value computed is not used"
	case "bar":
	}

	var z bool
	switch {
	case z:
		labelname:	// ERROR "missing statement after label"
	case false:
	}
}

func g() {
	var z bool
	switch {
	case z:
		labelname:	// ERROR "label labelname defined and not used|previous definition|defined and not used"
	}

	switch {
	case z:
		labelname: ;	// ERROR "label labelname already defined at LINE-5|label .*labelname.* already defined"
	case false:
	}
}

// GnoError:
// line 24: illegal label declaration
// line 25: expected ';', found 'case'
// line 30: illegal label declaration
// line 37: illegal label declaration
// line 38: expected ';', found 'case'
// line 43: illegal label declaration
// line 51: expected statement, found 'case'
// line 59: labelname redeclared in this block
// 	previous declaration at issue14006.go:50:3
// line 60: label labelname defined and not used
// line 64: labelname redeclared in this block
// 	previous declaration at issue14006.go:50:3
// line 67: expected ';', found 'EOF'

// GnoWordingDiffers:
// line 24: gc expects: unexpected :|expected .*;.* or .*}.* or newline|value computed is not used; gno: illegal label declaration
// line 30: gc expects: unexpected :|expected .*;.* or .*}.* or newline|value computed is not used; gno: illegal label declaration
// line 37: gc expects: unexpected :|expected .*;.* or .*}.* or newline|value computed is not used; gno: illegal label declaration
// line 43: gc expects: unexpected :|expected .*;.* or .*}.* or newline|value computed is not used; gno: illegal label declaration
// line 59: gc expects: label labelname defined and not used|previous definition|defined and not used; gno: labelname redeclared in this block
// 	previous declaration at issue14006.go:50:3
// line 64: gc expects: label labelname already defined at LINE-5|label .*labelname.* already defined; gno: labelname redeclared in this block
// 	previous declaration at issue14006.go:50:3

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 50; over-strict: lines 25, 38, 51, 60, 67

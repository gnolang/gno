// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f() (_ int, err error) {
	return
}

func g() (x int, _ error) {
	return
}

func h() (_ int, _ error) {
	return
}

func i() (int, error) {
	return // ERROR "not enough return values|not enough arguments to return"
}

func f1() (_ int, err error) {
	return 1, nil
}

func g1() (x int, _ error) {
	return 1, nil
}

func h1() (_ int, _ error) {
	return 1, nil
}

func ii() (int, error) {
	return 1, nil
}

// GnoError:
// line 21: 2: [function "i" does not terminate]
// line 22: expected 2 return values
// line 23: expected declaration, found '}'

// GoTypeCheckError:
// line 22: not enough return values
// 	have ()
// 	want (int, error)

// GnoOverStrictError:
// line 21: 2: [function "i" does not terminate]
// line 23: expected declaration, found '}'

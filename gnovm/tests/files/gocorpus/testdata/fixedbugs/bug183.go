// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T int

func f() {
	var x struct { T };
	var y struct { T T };
	x = y;	// ERROR "cannot|incompatible"
	_ = x;
}

type T1 struct { T }
type T2 struct { T T }

func g() {
	var x T1;
	var y T2;
	x = y;	// ERROR "cannot|incompatible"
	_ = x;
}

// GnoError:
// line 24: cannot use main.T2 as main.T1 without explicit conversion

// GoTypeCheckError:
// line 14: cannot use y (variable of type struct{T T}) as struct{T} value in assignment
// line 24: cannot use y (variable of struct type T2) as T1 value in assignment

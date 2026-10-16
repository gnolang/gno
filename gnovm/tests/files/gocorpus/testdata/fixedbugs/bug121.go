// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T func()

type I interface {
	f, g ();  // ERROR "unexpected comma"
}

type J interface {
	h T;  // ERROR "syntax|signature"
}

// GnoError:
// line 12: expected ';', found ','

// ParseBlocked:
// line 16: gc expects: syntax|signature

// GnoWordingDiffers:
// line 12: gc expects: unexpected comma; gno: expected ';', found ','

// Verdict: Clean errorcheck

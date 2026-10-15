// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type P *struct{}

func _() {
	// want an error even if the switch is empty
	var a struct{ _ func() }
	switch a /* ERROR "cannot switch on a" */ {
	}

	switch a /* ERROR "cannot switch on a" */ {
	case a: // no follow-on error here
	}

	// this is ok because f can be compared to nil
	var f func()
	switch f {
	}

	switch f {
	case nil:
	}

	switch (func())(nil) {
	case nil:
	}

	switch (func())(nil) {
	case f /* ERRORx `invalid case f in switch on .* \(func can only be compared to nil\)` */ :
	}

	switch nil /* ERROR "use of untyped nil in switch expression" */ {
	}

	// this is ok
	switch P(nil) {
	case P(nil):
	}
}

// GnoError:
// line 15: 3: struct{_ func()} is not comparable
// line 32: 3: func() can only be compared to nil

// GnoWordingDiffers:
// line 15: gc expects: cannot switch on a; gno: 3: struct{_ func()} is not comparable
// line 33: gc expects: invalid case f in switch on .* \(func can only be compared to nil\); gno: 3: func() can only be compared to nil

// Verdict: KnownIssue errorcheck: #6344; uncaught markers: lines 12, 36

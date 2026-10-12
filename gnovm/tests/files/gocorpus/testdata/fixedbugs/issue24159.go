// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type intAlias = int

func f() {
	switch interface{}(nil) {
	case uint8(0):
	case byte(0): // ERROR "duplicate case"
	case int32(0):
	case rune(0): // ERROR "duplicate case"
	case int(0):
	case intAlias(0): // ERROR "duplicate case"
	}
}

// GoTypeCheckError:
// line 13
// line 14
// line 15
// line 16
// line 17
// line 18

// GoTypesOnly:
// line 14: gc expects: duplicate case
// line 16: gc expects: duplicate case
// line 18: gc expects: duplicate case

// Verdict: Clean errorcheck

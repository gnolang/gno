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
// line 14: duplicate case byte(0) (constant 0 of type byte) in expression switch
// line 16: duplicate case rune(0) (constant 0 of type rune) in expression switch
// line 18: duplicate case intAlias(0) (constant 0 of int type intAlias) in expression switch

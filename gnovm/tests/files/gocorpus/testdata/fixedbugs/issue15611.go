// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

// These error messages are for the invalid literals on lines 19 and 20:

// ERROR "newline in character literal|newline in rune literal"
// ERROR "invalid character literal \(missing closing '\)|rune literal not terminated"

const (
	_ = ''     // ERROR "empty character literal or unescaped ' in character literal|empty rune literal"
	_ = 'f'
	_ = 'foo'  // ERROR "invalid character literal \(more than one character\)|more than one character in rune literal"
//line issue15611.go:11
	_ = '
	_ = '

// GnoError:
// line 10: rune literal not terminated (and 4 more errors)

// GoTypeCheckError:
// line 15: illegal rune literal (and 4 more errors)

// GnoOverStrictError:
// line 10: rune literal not terminated (and 4 more errors)

// UncaughtError:
// line 11: uncaught; gc expects: newline in character literal|newline in rune literal
// line 12: uncaught; gc expects: 
// line 17: uncaught; gc expects: 

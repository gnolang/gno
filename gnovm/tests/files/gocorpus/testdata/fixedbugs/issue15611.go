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

// GnoStaticIncomplete: covered 1 of 4 markers (Gno preprocess: 0, go/types guard: 1); Gno's own preprocess flags none (lenient); the rest are caught by neither — a runnable variant may exercise more

// GnoOverStrictError:
// line 10: rune literal not terminated (and 4 more errors)

// GoTypeCheckError:
// line 15: illegal rune literal (and 4 more errors)

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)

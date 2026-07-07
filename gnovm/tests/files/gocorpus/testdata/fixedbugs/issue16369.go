// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T interface { // ERROR "invalid recursive type: anonymous interface refers to itself"
	M(interface {
		T
	})
}

// GnoError:
// line 9: 2: invalid recursive type: T -> T
// line 10: expected declaration, found M
// line 11: expected declaration, found T
// line 12: expected declaration, found '}'
// line 13: expected declaration, found '}'

// GnoOverStrictError:
// line 10: expected declaration, found M
// line 11: expected declaration, found T
// line 12: expected declaration, found '}'
// line 13: expected declaration, found '}'

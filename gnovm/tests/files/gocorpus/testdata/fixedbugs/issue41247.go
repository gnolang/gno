// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f() [2]int {
	return [...]int{2: 0} // ERROR "cannot use \[\.\.\.\]int{.*} \(.*type \[3\]int\)"
}

// GnoError:
// line 9: 2: [function "f" does not terminate]
// line 10: cannot use [3]int as [2]int
// line 11: expected declaration, found '}'

// GoTypeCheckError:
// line 10: cannot use [...]int{…} (value of type [3]int) as [2]int value in return statement

// GnoOverStrictError:
// line 9: 2: [function "f" does not terminate]
// line 11: expected declaration, found '}'

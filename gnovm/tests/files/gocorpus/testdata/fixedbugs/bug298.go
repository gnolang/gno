// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ddd

func Sum() int
	for i := range []int{} { return i }  // ERROR "statement outside function|expected"

// GnoError:
// line 9: function Sum does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 10: expected declaration, found 'for'

// GoTypeCheckError:
// line 10: expected declaration, found 'for'

// GnoOverStrictError:
// line 9: function Sum does not have a body but is not natively defined (did you build after pulling from the repository?)

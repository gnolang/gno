// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	switch main() := interface{}(nil).(type) {	// ERROR "invalid variable name|cannot use .* as value"
	default:
	}
}

// GnoError:
// line 10: no new variables on left side of :=
// line 11: expected '}', found 'default'
// line 13: expected declaration, found '}'

// GoTypeCheckError:
// line 10: invalid syntax tree: incorrect form of type switch guard

// GnoOverStrictError:
// line 11: expected '}', found 'default'
// line 13: expected declaration, found '}'

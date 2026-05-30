// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

const A = complex(0()) // ERROR "cannot call non-function"

// GnoError:
// line 9: name complex not defined in fileset with files [issue17038.go]

// GoTypeCheckError:
// line 9: invalid operation: cannot call 0 (untyped int constant): untyped int is not a function

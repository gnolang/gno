// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 2451, 2452 
package foo

func f() error { return 0 } // ERROR "cannot use 0 (.type int.)?|has no methods"

func g() error { return -1 }  // ERROR "cannot use -1 (.type int.)?|has no methods"

// GnoError:
// line 10: <untyped> bigint does not implement .uverse.error (missing method Error)
// line 12: <untyped> bigint does not implement .uverse.error (missing method Error)

// GoTypeCheckError:
// line 10: cannot use 0 (constant of type int) as error value in return statement: int does not implement error (missing method Error)
// line 12: cannot use -1 (constant of type int) as error value in return statement: int does not implement error (missing method Error)

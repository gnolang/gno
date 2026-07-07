// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
import os "os"
type _ os.FileInfo
func f() (os int) {
	 // In the next line "os" should refer to the result variable, not
	 // to the package.
	 v := os.Open("", 0, 0);	// ERROR "undefined"
	 _ = v
	 return 0
}

// GnoOverStrictError:
// line 9: name FileInfo not declared

// GoTypeCheckError:
// line 13: os.Open undefined (type int has no field or method Open)

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)

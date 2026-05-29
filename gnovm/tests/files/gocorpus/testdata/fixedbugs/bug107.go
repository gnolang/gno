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

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them
// GnoError:
// line 9: name FileInfo not declared

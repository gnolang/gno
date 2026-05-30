// errorcheck -0 -l -d=wb

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test write barrier for implicit assignments to result parameters
// that have escaped to the heap.

package issue13587

import "errors"

func escape(p *error)

func F() (err error) {
	escape(&err)
	return errors.New("error") // ERROR "write barrier"
}

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// KnownIssue:
// line 14: function escape does not have a body but is not natively defined (did you build after pulling from the repository?)

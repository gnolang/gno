// errorcheck -std
// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Issue 18331: We should catch invalid pragma verbs
// for code that resides in the standard library.
package issue18331

//go:unknown // ERROR "//go:unknown is not allowed in the standard library"
func foo()

//go:nowritebarrierc // ERROR "//go:nowritebarrierc is not allowed in the standard library"
func bar()

//go:noesape // ERROR "//go:noesape is not allowed in the standard library"
func groot()

//go:noescape
func hey() { // ERROR "can only use //go:noescape with external func implementations"
}

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them
// GnoError:
// line 10: function foo does not have a body but is not natively defined (did you build after pulling from the repository?)

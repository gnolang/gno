// run

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 8132. stack walk handling of panic stack was confused
// about what was legal.

package main

import "runtime"

var p *int

func main() {
	func() {
		defer func() {
			runtime.GC()
			recover()
		}()
		var x [8192]byte
		func(x [8192]byte) {
			defer func() {
				if err := recover(); err != nil {
					println(*p)
				}
			}()
			println(*p)
		}(x)
	}()
}

// GnoOutput:

// GnoError:
// runtime error: invalid memory address or nil pointer dereference

// GoOutput:

// KnownIssue:
// TODO: explain the Gno bug (Gno errors where Go runs clean)

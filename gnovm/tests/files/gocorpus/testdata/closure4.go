// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check that calling a nil func causes a proper panic.

package main

func main() {
	defer func() {
		err := recover()
		if err == nil {
			panic("panic expected")
		}
	}()

	var f func()
	f()
}

// GnoOutput:

// GnoError:
// runtime error: invalid memory address or nil pointer dereference

// GoOutput:

// KnownIssue:
// TODO: explain the Gno bug (Gno errors where Go runs clean)

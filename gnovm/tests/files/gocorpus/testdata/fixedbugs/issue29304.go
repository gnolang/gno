// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that relocation target go.builtin.error.Error
// is defined and the code links and runs correctly.

package main

import "errors"

func main() {
	err := errors.New("foo")
	if error.Error(err) != "foo" {
		panic("FAILED")
	}
}


// Tracked: issue #5787 (method expressions: interface/promoted/mixed-receiver forms); broken on master, no PR yet.

// GnoOutput:

// GnoError:
// main/issue29304.go:16:5-16: unknown *DeclaredType method named Error

// GoOutput:

// KnownIssue:
// Method expressions on interface types are unsupported: error.Error(err)
// is rejected at preprocess ("unknown *DeclaredType method named Error")
// instead of yielding a func with the receiver as first parameter.

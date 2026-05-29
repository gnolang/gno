// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "os"

var _ = os.Open // avoid imported and not used error

type T struct {
	File int
}

func main() {
	_ = T{
		os.File: 1, // ERROR "invalid field name os.File|unknown field"
	}
}

// GoTypeCheckError:
// line 19: invalid field name os.File in struct literal

// KnownIssue:
// line 11: name Open not declared

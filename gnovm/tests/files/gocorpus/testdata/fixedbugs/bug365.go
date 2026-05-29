// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// check that compiler doesn't stop reading struct def
// after first unknown type.

// Fixes issue 2110.

package main

type S struct {
	err foo.Bar // ERROR "undefined|expected package"
	Num int
}

func main() {
	s := S{}
	_ = s.Num // no error here please
}

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them
// GnoError:
// line 14: 2: name foo not defined in fileset with files [bug365.go]

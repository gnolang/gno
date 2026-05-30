// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	// 6g used to compile these as two different
	// hash codes so it missed the duplication
	// and worse, compiled the wrong code
	// for one of them.
	var x interface{};
	switch x.(type) {
	case func(int):
	case func(f int):	// ERROR "duplicate"
	}
}

// GoTypeCheckError:
// line 17: duplicate case func(f int) in type switch

// KnownIssue:
// line 15: 3: duplicate type func(int) in type switch

// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	var buf [10]int;
	for ; len(buf); {  // ERROR "bool"
	}
}

/*
uetli:/home/gri/go/test/bugs gri$ 6g bug209.go
bug209.go:5: Bus error
*/

// GnoError:
// line 11: 3: expected typed bool kind, but got IntKind
// line 13: expected declaration, found '}'

// GoTypeCheckError:
// line 11: non-boolean condition in for statement

// GnoOverStrictError:
// line 13: expected declaration, found '}'

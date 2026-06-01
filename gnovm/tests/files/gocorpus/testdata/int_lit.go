// run

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test integer literal syntax.

package main

import "os"

func main() {
	s := 	0 +
		123 +
		0123 +
		0000 +
		0x0 +
		0x123 +
		0X0 +
		0X123
	if s != 788 {
		print("s is ", s, "; should be 788\n")
		os.Exit(1)
	}
}

// TypeCheckError:
// main/int_lit.go:24:6: undefined: os.Exit

// GnoOutput:

// GnoError:
// main/int_lit.go:24:3-10: name Exit not declared

// GoOutput:

// Unsupported: unsupported stdlib symbol in Gno: Exit

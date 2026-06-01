// run

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "os"

func main() {
	// This bug doesn't arise with [...]int, or []interface{} or [3]interface{}.
	a := [...]interface{} { 1, 2, 3 };
	n := 1;
	for _, v := range a {
		if v.(int) != n {
			println("BUG:", n, v.(int));
			os.Exit(0);
		}
		n++;
	}
}

// TypeCheckError:
// main/bug187.go:18:7: undefined: os.Exit

// GnoOutput:

// GnoError:
// main/bug187.go:18:4-11: name Exit not declared

// GoOutput:

// Unsupported: unsupported stdlib symbol in Gno: Exit

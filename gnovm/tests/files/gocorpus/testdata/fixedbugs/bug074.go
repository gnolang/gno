// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	x := string{'a', 'b', '\n'};	// ERROR "composite"
	print(x);
}

// GnoError:
// line 10: unexpected composite lit type string

// GoTypeCheckError:
// line 10: invalid composite literal type string

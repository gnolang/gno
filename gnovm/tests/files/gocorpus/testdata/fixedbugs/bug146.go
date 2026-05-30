// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	type Slice []byte;
	a := [...]byte{ 0 };
	b := Slice(a[0:]);	// This should be OK.
	c := Slice(a);		// ERROR "invalid|illegal|cannot"
	_, _ = b, c;
}

// GnoError:
// line 13: cannot convert a<VPBlock(1,1)> (of type [1]uint8) to type main[main/bug146.go:9:1-15:2].Slice

// GoTypeCheckError:
// line 13: cannot convert a (variable of type [1]byte) to type Slice

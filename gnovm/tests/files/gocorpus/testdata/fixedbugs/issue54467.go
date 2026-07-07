// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

func main() {
	var x [64]byte
	for i := range x {
		x[i] = byte(i)
	}
	y := x

	copy(x[4:36], x[2:34])
	*(*[32]byte)(y[4:36]) = *(*[32]byte)(y[2:34])

	for i := range x {
		if x[i] != y[i] {
			fmt.Printf("x[%v] = %v; y[%v] = %v\n", i, x[i], i, y[i])
		}
	}
}


// KnownDivergence:
// KnownDivegence:
// Go1.17 unsupport.

// GnoOutput:

// GnoError:
// main/issue54467.go:19:27-47: cannot convert y<VPBlock(1,1)>[(const (2 int)):(const (34 int))] (of type []uint8) to type *[32]uint8

// GoOutput:

// run

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test internal print routines that are generated
// by the print builtin.  This test is not exhaustive,
// we're just checking that the formatting is correct.

package main

func main() {
	println((interface{})(nil)) // printeface
	println((interface {        // printiface
		f()
	})(nil))
	println((map[int]int)(nil)) // printpointer
	println(([]int)(nil))       // printslice
	println(int64(-7))          // printint
	println(uint64(7))          // printuint
	println(uint32(7))          // printuint
	println(uint16(7))          // printuint
	println(uint8(7))           // printuint
	println(uint(7))            // printuint
	println(uintptr(7))         // printuint
	println(8.0)                // printfloat
	println(complex(9.0, 10.0)) // printcomplex
	println(true)               // printbool
	println(false)              // printbool
	println("hello")            // printstring
	println("one", "two")       // printsp

	// test goprintf
	defer println((interface{})(nil))
	defer println((interface {
		f()
	})(nil))
	defer println((map[int]int)(nil))
	defer println(([]int)(nil))
	defer println(int64(-11))
	defer println(uint64(12))
	defer println(uint32(12))
	defer println(uint16(12))
	defer println(uint8(12))
	defer println(uint(12))
	defer println(uintptr(12))
	defer println(13.0)
	defer println(complex(14.0, 15.0))
	defer println(true)
	defer println(false)
	defer println("hello")
	defer println("one", "two")
}

// GnoOutput:

// GnoError:
// main/print.go:26:10-17: name uintptr not declared

// GoOutput:
// (0x0,0x0)
// (0x0,0x0)
// 0x0
// [0/0]0x0
// -7
// 7
// 7
// 7
// 7
// 7
// 7
// +8.000000e+000
// (+9.000000e+000+1.000000e+001i)
// true
// false
// hello
// one two
// one two
// hello
// false
// true
// (+1.400000e+001+1.500000e+001i)
// +1.300000e+001
// 12
// 12
// 12
// 12
// 12
// 12
// -11
// [0/0]0x0
// 0x0
// (0x0,0x0)
// (0x0,0x0)

// KnownIssue:
// TODO: explain the Gno bug (Gno errors where Go runs clean)

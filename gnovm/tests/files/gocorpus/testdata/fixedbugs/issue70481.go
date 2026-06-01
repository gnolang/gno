// run

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

const maxUint64 = (1 << 64) - 1

//go:noinline
func f(n uint64) uint64 {
	return maxUint64 - maxUint64%n
}

func main() {
	for i := uint64(1); i < 20; i++ {
		println(i, maxUint64-f(i))
	}
}

// GnoOutput:
// 1 0
// 2 1
// 3 0
// 4 3
// 5 0
// 6 3
// 7 1
// 8 7
// 9 6
// 10 5
// 11 4
// 12 3
// 13 2
// 14 1
// 15 0
// 16 15
// 17 0
// 18 15
// 19 16

// GoOutput:
// 1 0
// 2 1
// 3 0
// 4 3
// 5 0
// 6 3
// 7 1
// 8 7
// 9 6
// 10 5
// 11 4
// 12 3
// 13 2
// 14 1
// 15 0
// 16 15
// 17 0
// 18 15
// 19 16

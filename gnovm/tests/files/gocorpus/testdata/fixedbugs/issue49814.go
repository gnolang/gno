// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// "must be integer" error is for 32-bit architectures
type V [1 << 50]byte // ERROR "larger than address space|invalid array length"

var X [1 << 50]byte // ERROR "larger than address space|invalid array length"

func main() {}

// Unsupported: Gno rejects the 1<<50 array at runtime (makeslice: len out of range), not at compile time like gc; no compile-time error to pin.

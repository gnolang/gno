// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main()
func main /* ERROR "no arguments and no return values" */ /* ERROR "redeclared" */ (int)
func main /* ERROR "no arguments and no return values" */ /* ERROR "redeclared" */ () int

// GnoError:
// line 8: main redeclared in this block
// 	previous declaration at main0.go:7:6
// line 9: main redeclared in this block
// 	previous declaration at main0.go:7:6

// Clean: errorcheck

// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main(int)  {}           // ERROR "func main must have no arguments and no return values"
func main() int { return 1 } // ERROR "func main must have no arguments and no return values" "main redeclared in this block"

func init(int)  {}           // ERROR "func init must have no arguments and no return values"
func init() int { return 1 } // ERROR "func init must have no arguments and no return values"

// GnoError:
// line 7: 29: wrong argument count in call to init.1<VPBlock(2,1)>
// line 10: main redeclared in this block
// 	previous declaration at mainsig.go:9:6

// GoTypeCheckError:
// line 13: func init must have no arguments and no return values

// GnoOverStrictError:
// line 7: 29: wrong argument count in call to init.1<VPBlock(2,1)>

// UncaughtError:
// line 9: uncaught; gc expects: func main must have no arguments and no return values
// line 12: uncaught; gc expects: func init must have no arguments and no return values

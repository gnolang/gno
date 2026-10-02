package main

// A name is addressable only when it denotes a variable. These all do,
// including variables that merely hold a func, so they keep working.

import "strconv"

var gvar int

var gfn = func() int { return 1 }

func fn() int { return 2 }

func main() {
	var local func() int
	assigned := fn
	_ = &gvar
	_ = &gfn
	_ = &local
	_ = &assigned
	// The same rule across a package boundary: ErrRange is a variable, so
	// its address is legal, while &strconv.Atoi is not (addressable_12c).
	_ = &strconv.ErrRange
	println("ok")
}

// Output:
// ok

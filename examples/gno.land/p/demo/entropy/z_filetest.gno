package main

import (
	"std"

	"gno.land/p/demo/entropy"
)

func main() {
	// initial
	println("---")
	r := entropy.New()
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())

	// should be the same
	println("---")
	r = entropy.New()
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())

	std.TestSkipHeights(1)
	println("---")
	r = entropy.New()
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
}

// Output:
// ---
// 4129293727
// 2141104956
// 1950222777
// 3348280598
// 438354259
// ---
// 4129293727
// 2141104956
// 1950222777
// 3348280598
// 438354259
// ---
// 49506731
// 1539580078
// 2695928529
// 1895482388
// 3462727799

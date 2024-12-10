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
	println(r.Value64())

	// should be the same
	println("---")
	r = entropy.New()
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value64())

	std.TestSkipHeights(1)
	println("---")
	r = entropy.New()
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value())
	println(r.Value64())
}

// Output:
// ---
// 4129293727
// 2141104956
// 1950222777
// 3348280598
// 438354259
// 6353385488959065197
// ---
// 4129293727
// 2141104956
// 1950222777
// 3348280598
// 438354259
// 6353385488959065197
// ---
// 49506731
// 1539580078
// 2695928529
// 1895482388
// 3462727799
// 16745038698684748445

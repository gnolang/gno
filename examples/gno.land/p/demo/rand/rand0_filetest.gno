package main

import (
	"std"

	"gno.land/p/demo/rand"
)

func main() {
	// initial
	println("---")
	r := rand.New()
	println(r.Intn(1000))
	println(r.Intn(1000))
	println(r.Intn(1000))
	println(r.Intn(1000))
	println(r.Intn(1000))

	// should be the same
	println("---")
	r = rand.New()
	println(r.Intn(1000))
	println(r.Intn(1000))
	println(r.Intn(1000))
	println(r.Intn(1000))
	println(r.Intn(1000))

	std.TestSkipHeights(1)
	println("---")
	r = rand.New()
	println(r.Intn(1000))
	println(r.Intn(1000))
	println(r.Intn(1000))
	println(r.Intn(1000))
	println(r.Intn(1000))
}

// Output:
// ---
// 777
// 257
// 74
// 177
// 802
// ---
// 777
// 257
// 74
// 177
// 802
// ---
// 269
// 233
// 591
// 936
// 908

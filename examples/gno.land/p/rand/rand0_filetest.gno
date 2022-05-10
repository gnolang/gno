package main

import (
	"std"

	"gno.land/p/rand"
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
// 71
// 908
// 897
// 955
// 61
// ---
// 71
// 908
// 897
// 955
// 61
// ---
// 410
// 919
// 51
// 342
// 683

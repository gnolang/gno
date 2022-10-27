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
// 426
// 168
// 493
// 512
// 463
// ---
// 426
// 168
// 493
// 512
// 463
// ---
// 333
// 673
// 723
// 699
// 444

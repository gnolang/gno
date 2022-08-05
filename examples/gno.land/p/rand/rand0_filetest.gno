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
// 571
// 638
// 608
// 851
// 455
// ---
// 571
// 638
// 608
// 851
// 455
// ---
// 989
// 704
// 782
// 367
// 888

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
// 343
// 682
// 557
// 741
// 503
// ---
// 343
// 682
// 557
// 741
// 503
// ---
// 790
// 699
// 118
// 747
// 755

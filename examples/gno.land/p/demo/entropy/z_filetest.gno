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
// 1246803619
// 1485738466
// 909405889
// 3591022912
// 1667505503
// ---
// 1246803619
// 1485738466
// 909405889
// 3591022912
// 1667505503
// ---
// 1461983919
// 884213588
// 1655111641
// 2138224702
// 396911747

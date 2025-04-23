package main

import (
	"testing"

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

	testing.SkipHeights(1)
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
// 2541982999
// 3833705520
// 3869786793
// 3937354882
// 2032680379
// 16793415398083372621
// ---
// 2541982999
// 3833705520
// 3869786793
// 3937354882
// 2032680379
// 16793415398083372621
// ---
// 731521123
// 3474881922
// 657635905
// 2551591584
// 3290088095
// 17335698837474056573

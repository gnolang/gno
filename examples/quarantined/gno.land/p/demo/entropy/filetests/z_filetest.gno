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
// 2916840437
// 3981870829
// 1311573733
// 2267403229
// 484167125
// 5277608318285564101
// ---
// 2916840437
// 3981870829
// 1311573733
// 2267403229
// 484167125
// 5277608318285564101
// ---
// 3603434145
// 911682639
// 1320246589
// 2861798763
// 1938803929
// 17792872788806539637

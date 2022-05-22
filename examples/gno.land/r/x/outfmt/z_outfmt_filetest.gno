package main

import (
	"std"

	"gno.land/r/x/outfmt"
)

func main() {
	println(outfmt.Render("/"))
	println("done")
}

// Output:
// 10100000000
// 10000000000
// 100000000
// 0
// 0
// done

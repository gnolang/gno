package main

import (
	"gno.land/p/demo/entropy"
	"gno.land/r/demo/games/ga"
)

func main() {

	r := entropy.New()

	gen, target := ga.Start(r)
}

// Output:
// Target reached!

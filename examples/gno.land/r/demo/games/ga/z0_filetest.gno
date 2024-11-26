package main

import (
	"std"

	"gno.land/p/demo/entropy"
	"gno.land/r/demo/games/ga"
	"strconv"
)

func main() {

	r := entropy.New()

	gen, target := ga.Start(r)
	//println("found " + target + " in " + strconv.Itoa(gen) + " generation")
}

// Output:
// Target reached!

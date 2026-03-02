package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	bid, found := boards2.GetBoardIDFromName(cross, "foobar")

	println(found)
	println(bid == 0)
}

// Output:
// false
// true

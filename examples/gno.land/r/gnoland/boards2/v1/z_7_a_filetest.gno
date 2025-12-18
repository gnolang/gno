package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	name          = "test123"
)

var bid boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, name, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	bid2, found := boards2.GetBoardIDFromName(cross, name)

	println(found)
	println(bid2 == bid)
}

// Output:
// true
// true

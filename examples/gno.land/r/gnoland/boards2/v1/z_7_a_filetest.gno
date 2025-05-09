package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	name  = "test123"
)

var bid boards2.BoardID

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)(name)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	bid2, found := cross(boards2.GetBoardIDFromName)(name)

	println(found)
	println(bid2 == bid)
}

// Output:
// true
// true

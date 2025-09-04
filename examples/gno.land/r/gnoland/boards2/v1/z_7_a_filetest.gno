package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	name  = "test123"
)

var bid boards2.BoardID

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, name, false)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	bid2, found := boards2.GetBoardIDFromName(cross, name)

	println(found)
	println(bid2 == bid)
}

// Output:
// true
// true

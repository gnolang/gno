package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid boards2.BoardID
	tid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
	tid = boards2.CreateThread(cross, bid, "foo", "bar")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.FreezeThread(cross, bid, tid)

	println(boards2.IsThreadFrozen(bid, tid))
}

// Output:
// true

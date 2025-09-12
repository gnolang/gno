package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid boards2.BoardID
	tid boards2.PostID
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
	tid = boards2.CreateThread(cross, bid, "foo", "bar")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.FreezeThread(cross, bid, tid)

	println(boards2.IsThreadFrozen(bid, tid))
}

// Output:
// true

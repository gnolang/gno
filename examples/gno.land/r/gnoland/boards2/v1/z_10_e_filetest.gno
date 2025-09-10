package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	pid = boards2.CreateThread(cross, bid, "Foo", "bar")
	boards2.FlagThread(cross, bid, pid, "")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.FlagThread(cross, bid, pid, "")
}

// Error:
// item flag count threshold exceeded: 1

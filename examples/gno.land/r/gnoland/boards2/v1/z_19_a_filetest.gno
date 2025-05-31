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
	bid = cross(boards2.CreateBoard)("test123", false)
	tid = cross(boards2.CreateThread)(bid, "foo", "bar", false)

	cross(boards2.FreezeBoard)(bid)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	// Attempt to freeze a thread on frozen board
	cross(boards2.FreezeThread)(bid, tid)
}

// Error:
// board is frozen

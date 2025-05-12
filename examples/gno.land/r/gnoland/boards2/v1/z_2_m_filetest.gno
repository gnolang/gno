package main

import (
	"std"
	"strings"
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
	bid = cross(boards2.CreateBoard)("test-board", false)
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.CreateReply)(bid, tid, 0, strings.Repeat("x", boards2.MaxReplyLength+1))
}

// Error:
// reply is too long, maximum allowed is 300 characters

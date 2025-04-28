package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var bid boards2.BoardID

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test123")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.CreateReply)(bid, 404, 0, "comment")
}

// Error:
// thread does not exist with ID: 404

package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	srcBID = cross(boards2.CreateBoard)("src-board")
	dstBID = cross(boards2.CreateBoard)("dst-board")

	srcTID = cross(boards2.CreateThread)(srcBID, "Foo", "bar")

	testing.SetOriginCaller(user)
}

func main() {
	cross(boards2.CreateRepost)(srcBID, srcTID, "foo", "bar", dstBID)
}

// Error:
// unauthorized

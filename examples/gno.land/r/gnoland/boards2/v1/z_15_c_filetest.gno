package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var (
	srcBID boards.ID
	dstBID boards.ID
	srcTID boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	srcBID = boards2.CreateBoard(cross, "src-board", false)
	dstBID = boards2.CreateBoard(cross, "dst-board", false)

	srcTID = boards2.CreateThread(cross, srcBID, "Foo", "bar")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.CreateRepost(cross, srcBID, srcTID, dstBID, "foo", "bar")
}

// Error:
// unauthorized

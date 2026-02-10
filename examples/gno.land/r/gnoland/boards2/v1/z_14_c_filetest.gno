package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var bid boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.DeleteReply(cross, bid, 404, 1)
}

// Error:
// thread does not exist with ID: 404

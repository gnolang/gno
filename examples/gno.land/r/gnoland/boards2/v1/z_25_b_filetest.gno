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
	bid = boards2.CreateBoard(cross, "test123", false, false)
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))

	boards2.RequestInvite(cross, bid)
}

// Error:
// caller must be user

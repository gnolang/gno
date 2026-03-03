package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

var bid boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, false)
	boards2.FreezeBoard(cross, bid)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.FreezeBoard(cross, bid)
}

// Error:
// board is frozen

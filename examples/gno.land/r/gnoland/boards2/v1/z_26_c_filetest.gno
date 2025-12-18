package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	user  address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
)

var bid boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.AcceptInvite(cross, bid, user)
}

// Error:
// invite request not found

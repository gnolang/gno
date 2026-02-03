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
	bid = boards2.CreateBoard(cross, "test123", false, false)

	testing.SetRealm(testing.NewUserRealm(user))
	boards2.RequestInvite(cross, bid)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.AcceptInvite(cross, bid, user)

	println(boards2.IsMember(bid, user))
	println()
	println(boards2.Render("test123/invites"))
}

// Output:
// true
//
// # test123 Invite Requests
// ### Board has no invite requests

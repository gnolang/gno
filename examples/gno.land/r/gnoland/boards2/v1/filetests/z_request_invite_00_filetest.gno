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
}

func main() {
	testing.SetRealm(testing.NewUserRealm("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"))

	boards2.RequestInvite(cross, bid)

	println(boards2.Render("test123/invites"))
}

// Output:
// # test123 Invite Requests
// ### These users have requested to be invited to the board
// | User | Request Date | Actions |
// | --- | --- | --- |
// | [g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5](/u/g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5) | 2009-02-13 11:31pm UTC | [accept](/r/gnoland/boards2/v1$help&func=AcceptInvite&boardID=1&user=g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5) • [revoke](/r/gnoland/boards2/v1$help&func=RevokeInvite&boardID=1&user=g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5) |

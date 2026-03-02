package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	user  address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
)

var bid boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, false)

	// Request an invite as a user that is not a member
	testing.SetRealm(testing.NewUserRealm(user))
	boards2.RequestInvite(cross, bid)

	// Add user as a member idependently of the invite request
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.InviteMember(cross, bid, user, boards2.RoleGuest)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.AcceptInvite(cross, bid, user)
}

// Error:
// user is already a member

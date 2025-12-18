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

	// Invite the user as a moderator
	boards2.InviteMember(cross, bid, user, boards2.RoleModerator)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.Ban(cross, bid, user, boards2.BanDay, "Reason")
}

// Error:
// owner, admin and moderator banning is not allowed

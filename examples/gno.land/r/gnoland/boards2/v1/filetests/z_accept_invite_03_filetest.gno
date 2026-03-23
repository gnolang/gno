package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
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
	// Caller is not a member and has no permission to invite
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.AcceptInvite(cross, bid, user)
}

// Error:
// unauthorized

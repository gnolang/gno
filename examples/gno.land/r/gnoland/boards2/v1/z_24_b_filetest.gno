package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	admin = std.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var bid boards2.BoardID

func init() {
	testing.SetRealm(std.NewUserRealm(owner))

	bid = boards2.CreateBoard(cross, "test123", false)
	boards2.InviteMember(cross, bid, admin, boards2.RoleAdmin)
}

func main() {
	testing.SetRealm(std.NewUserRealm(admin))

	// Calling as a non owner member should fail
	boards2.GetBoard(bid)
}

// Error:
// forbidden

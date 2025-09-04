package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	admin = chain.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var bid boards2.BoardID

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	bid = boards2.CreateBoard(cross, "test123", false)
	boards2.InviteMember(cross, bid, admin, boards2.RoleAdmin)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(admin))

	// Calling as a non owner member should fail
	boards2.GetBoard(bid)
}

// Error:
// forbidden

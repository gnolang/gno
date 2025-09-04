package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = chain.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)

	boards2.InviteMember(cross, bid, user, boards2.RoleGuest)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.RemoveMember(cross, bid, user)

	// Check that user is not a member
	println(boards2.IsMember(bid, user))
}

// Output:
// false

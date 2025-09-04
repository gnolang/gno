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
	bid   = boards2.BoardID(0)                                        // Operate on realm DAO instead of individual boards
	role  = boards2.RoleOwner
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	boards2.InviteMember(cross, bid, user, role)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.InviteMember(cross, bid, user, role)
}

// Error:
// user is already a member

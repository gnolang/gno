package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	testing.SetOriginCaller(owner)
	bid = boards2.CreateBoard("test123")

	boards2.InviteMember(bid, admin, boards2.RoleGuest)
}

func main() {
	boards2.ChangeMemberRole(bid, "invalid address", boards2.RoleModerator)
}

// Error:
// invalid member address

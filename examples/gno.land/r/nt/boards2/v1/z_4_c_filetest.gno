package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	owner2 = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	admin  = std.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test123")

	boards2.InviteMember(bid, owner2, boards2.RoleOwner)
	boards2.InviteMember(bid, admin, boards2.RoleAdmin)

	std.TestSetOriginCaller(admin)
}

func main() {
	boards2.ChangeMemberRole(bid, owner2, boards2.RoleAdmin)
}

// Error:
// admins are not allowed to remove the Owner role

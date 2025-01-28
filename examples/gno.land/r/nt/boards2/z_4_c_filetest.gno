package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const (
	owner  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	owner2 = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	admin  = std.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
	bid    = boards2.BoardID(0) // Operate on realm DAO members instead of individual boards
)

func init() {
	std.TestSetOrigCaller(owner)
	boards2.InviteMember(bid, owner2, boards2.RoleOwner)
	boards2.InviteMember(bid, admin, boards2.RoleAdmin)

	std.TestSetOrigCaller(admin)
}

func main() {
	boards2.ChangeMemberRole(bid, owner2, boards2.RoleAdmin)
}

// Error:
// admins are not allowed to remove the Owner role

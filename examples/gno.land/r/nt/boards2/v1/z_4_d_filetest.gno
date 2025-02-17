package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	admin  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	admin2 = std.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("test123")

	boards2.InviteMember(bid, admin, boards2.RoleAdmin)
	boards2.InviteMember(bid, admin2, boards2.RoleAdmin)

	std.TestSetOrigCaller(admin)
}

func main() {
	boards2.ChangeMemberRole(bid, admin2, boards2.RoleOwner)
}

// Error:
// admins are not allowed to promote members to Owner

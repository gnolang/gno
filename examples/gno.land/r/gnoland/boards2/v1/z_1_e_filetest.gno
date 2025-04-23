package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	role  = boards2.RoleOwner
)

var bid boards2.BoardID

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("foo123") // Operate on board DAO members

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.InviteMember)(bid, user, role)

	// Check that user is invited
	println(boards2.HasMemberRole(0, user, role)) // Operate on realm DAO
	println(boards2.HasMemberRole(bid, user, role))
}

// Output:
// false
// true

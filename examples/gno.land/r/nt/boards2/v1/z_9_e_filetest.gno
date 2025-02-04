package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const (
	owner  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("test-board")
	pid = boards2.CreateThread(bid, "Foo", "bar")

	// Invite a member using a role with permission to delete threads
	boards2.InviteMember(bid, member, boards2.RoleAdmin)
	std.TestSetOrigCaller(member)
}

func main() {
	boards2.DeleteThread(bid, pid)

	// Ensure thread doesn't exist
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread does not exist with ID: 1

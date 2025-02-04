package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const (
	owner     = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	moderator = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("test-board")
	pid = boards2.CreateThread(bid, "Foo", "bar")

	// Invite a member using a role with permission to flag threads
	boards2.InviteMember(bid, moderator, boards2.RoleModerator)
	std.TestSetOrigCaller(moderator)
}

func main() {
	boards2.FlagThread(bid, pid, "")

	// Ensure that original thread content not visible
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread with ID: 1 has been flagged as inappropriate

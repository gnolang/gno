package main

import (
	"std"
	"strings"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner     = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	moderator = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid      boards2.BoardID
	rid, tid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	tid = boards2.CreateThread(bid, "Foo", "bar", false)
	rid = boards2.CreateReply(bid, tid, 0, "body")

	// Invite a member using a role with permission to flag replies
	boards2.InviteMember(bid, moderator, boards2.RoleModerator)
	std.TestSetOriginCaller(moderator)
}

func main() {
	boards2.FlagReply(bid, tid, rid, "")

	// Render content must contain a message about the hidden reply
	content := boards2.Render("test-board/1/2")
	println(strings.Contains(content, "\n> _Reply is hidden as it has been flagged as inappropriate_\n"))
}

// Output:
// true

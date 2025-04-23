package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "body")

	// Invite a member using a role with permission to delete replies
	cross(boards2.InviteMember)(bid, member, boards2.RoleAdmin)

	testing.SetOriginCaller(member)
}

func main() {
	cross(boards2.DeleteReply)(bid, tid, rid)

	// Ensure reply doesn't exist
	println(cross(boards2.Render)("test-board/1/2"))
}

// Output:
// Reply does not exist with ID: 2

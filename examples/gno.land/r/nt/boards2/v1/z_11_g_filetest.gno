package main

import (
	"std"
	"strings"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	title = "Test Thread"
	body  = "Test body"
	path  = "test-board/1"
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	pid = boards2.CreateThread(bid, "Foo", "bar", false)

	// Invite a member using a role with permission to edit threads
	boards2.InviteMember(bid, admin, boards2.RoleAdmin)
	std.TestSetOriginCaller(admin)
}

func main() {
	boards2.EditThread(bid, pid, title, body)

	// Render content must contains thread's title and body
	content := boards2.Render(path)
	println(strings.HasPrefix(content, "# "+title))
	println(strings.Contains(content, body))
}

// Output:
// true
// true

package main

import (
	"std"
	"strings"

	"gno.land/r/nt/boards2"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	title = "Test Thread"
	body  = "Test body"
	path  = "test-board/1"
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("test-board")
	pid = boards2.CreateThread(bid, "Foo", "bar")
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

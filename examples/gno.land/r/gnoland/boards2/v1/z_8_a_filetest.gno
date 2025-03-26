package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	title = "Test Thread"
	body  = "Test body"
	path  = "test-board/1"
)

var bid boards2.BoardID

func init() {
	testing.SetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
}

func main() {
	tid := boards2.CreateThread(bid, title, body)

	// Ensure that returned ID is right
	println(tid == 1)

	// Thread should not be frozen by default
	println(boards2.IsThreadFrozen(bid, tid))

	// Render content must contains thread's title and body
	content := boards2.Render(path)
	println(strings.HasPrefix(content, "## "+title))
	println(strings.Contains(content, body))
}

// Output:
// true
// false
// true
// true

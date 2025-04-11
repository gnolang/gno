package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid   boards2.BoardID
	pid   boards2.PostID
	title = strings.Repeat("X", boards2.MaxThreadTitleLength+1)
)

func init() {
	testing.SetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	pid = boards2.CreateThread(bid, "Foo", "bar")
}

func main() {
	boards2.EditThread(bid, pid, title, "bar")
}

// Error:
// thread title is too long, maximum allowed is 100 characters

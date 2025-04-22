package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid boards2.BoardID
	tid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.CreateReply)(bid, tid, 0, "> Markdown blockquote")
}

// Error:
// using Markdown headings, blockquotes or horizontal lines is not allowed in replies

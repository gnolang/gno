package main

import (
	"std"
	"strings"

	"gno.land/r/nt/boards2"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	srcBid boards2.BoardID
	dstBid boards2.BoardID
	srcTid boards2.PostID
)

func init() {
	std.TestSetOrigCaller(owner)
	srcBid = boards2.CreateBoard("src-board")
	dstBid = boards2.CreateBoard("dst-board")

	srcTid = boards2.CreateThread(srcBid, "Foo", "bar")
	std.TestSetOrigCaller(user)
}

func main() {
	_ = boards2.CreateRepost(srcBid, srcTid, "foo", "bar", dstBid)
}

// Error:
// unauthorized

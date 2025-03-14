package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	srcBID = boards2.CreateBoard("src-board")
	dstBID = boards2.CreateBoard("dst-board")

	srcTID = boards2.CreateThread(srcBID, "Foo", "bar")
	std.TestSetOriginCaller(user)
}

func main() {
	boards2.CreateRepost(srcBID, srcTID, "foo", "bar", dstBID)
}

// Error:
// unauthorized

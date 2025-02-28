package main

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	srcBid boards2.BoardID
	dstBid boards2.BoardID
	srcTid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	srcBid = boards2.CreateBoard("src-board")
	dstBid = boards2.CreateBoard("dst-board")

	srcTid = boards2.CreateThread(srcBid, "original title", "original text", false)
}

func main() {
	// Success case
	tid := boards2.CreateRepost(srcBid, srcTid, "repost title", "repost text", dstBid)
	p := ufmt.Sprintf("dst-board/%s", tid)
	out := boards2.Render(p)

	println(strings.Contains(out, "original text"))
	println(strings.Contains(out, "repost title"))
	println(strings.Contains(out, "repost text"))
}

// Output:
// true
// true
// true

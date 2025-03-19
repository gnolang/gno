package main

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	srcBID = boards2.CreateBoard("src-board")
	dstBID = boards2.CreateBoard("dst-board")

	srcTID = boards2.CreateThread(srcBID, "original title", "original text")
}

func main() {
	// Success case
	tID := boards2.CreateRepost(srcBID, srcTID, "repost title", "repost text", dstBID)
	p := ufmt.Sprintf("dst-board/%s", tID)
	out := boards2.Render(p)

	println(strings.Contains(out, "original text"))
	println(strings.Contains(out, "repost title"))
	println(strings.Contains(out, "repost text"))
}

// Output:
// true
// true
// true

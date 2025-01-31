package main

import (
	"std"
	"strings"

	"gno.land/r/nt/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	srcBid boards2.BoardID
	dstBid boards2.BoardID
	srcTid boards2.PostID
)

func init() {
	std.TestSetOrigCaller(owner)
	rootBid := boards2.CreateBoard("board1")
	rootTid := boards2.CreateThread(rootBid, "a", "original post content")
	bid := boards2.CreateBoard("board2")
	tid := boards2.CreateRepost(rootBid, rootTid, "b", "b", bid)

	srcBid = bid
	srcTid = tid
	dstBid = boards2.CreateBoard("board3")
}

func main() {
	// Sequential repost should work
	tid := boards2.CreateRepost(srcBid, srcTid, "c", "c", dstBid)
	boards2.Render("board3/1")
}

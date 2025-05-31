package main

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/ufmt"
	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	srcBID = cross(boards2.CreateBoard)("src-board", false)
	dstBID = cross(boards2.CreateBoard)("dst-board", false)

	srcTID = cross(boards2.CreateThread)(srcBID, "original title", "original text", false)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	// Success case
	tID := cross(boards2.CreateRepost)(srcBID, srcTID, "repost title", "repost text", dstBID)
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

package main

import (
	"chain/runtime"
	"strings"
	"testing"

	"gno.land/p/demo/ufmt"
	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	srcBID = boards2.CreateBoard(cross, "src-board", false)
	dstBID = boards2.CreateBoard(cross, "dst-board", false)

	srcTID = boards2.CreateThread(cross, srcBID, "original title", "original text")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	// Success case
	tID := boards2.CreateRepost(cross, srcBID, srcTID, "repost title", "repost text", dstBID)
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

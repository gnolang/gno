package main

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	"gno.land/p/nt/ufmt/v0"
	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

var (
	srcBID boards.ID
	dstBID boards.ID
	srcTID boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	srcBID = boards2.CreateBoard(cross, "src-board", false, false)
	dstBID = boards2.CreateBoard(cross, "dst-board", false, false)

	srcTID = boards2.CreateThread(cross, srcBID, "original title", "original text")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Success case
	tID := boards2.CreateRepost(cross, srcBID, srcTID, dstBID, "repost title", "repost text")
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

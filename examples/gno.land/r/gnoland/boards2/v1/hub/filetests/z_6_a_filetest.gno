// Test getting flags from an unflagged thread
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var (
	boardID  boards.ID
	threadID boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"))
	boardID = boards2.CreateBoard(cross, "test123", false, false)
	threadID = boards2.CreateThread(cross, boardID, "Title", "Body")
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	flags := hub.GetFlags(uint64(boardID), uint64(threadID), 0, 0, 1)
	println(len(flags))
}

// Output:
// 0

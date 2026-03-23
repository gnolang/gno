// Test getting reposts from a thread without reposts
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var thread hub.Thread

func init() {
	testing.SetRealm(testing.NewUserRealm("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"))
	boardID := boards2.CreateBoard(cross, "aaa123", false, false)
	threadID := boards2.CreateThread(cross, boardID, "Title", "Body")

	// Get readonly thread
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	thread, _ = hub.GetThread(uint64(boardID), uint64(threadID))
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	reposts := hub.GetReposts(thread.BoardID, thread.ID, 0, thread.RepostCount)
	println(len(reposts))
}

// Output:
// 0

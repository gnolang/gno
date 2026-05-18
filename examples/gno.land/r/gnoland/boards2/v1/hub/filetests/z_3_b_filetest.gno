// PKGPATH: gno.land/r/gnoland/boards2/v1/hub/filetests/z_3_b_filetest

// Test getting threads from a board
package z_3_b_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var board hub.Board

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID := boards2.CreateBoard(cross2(cur), "test123", false, false)

	// Create a couple of board threads
	boards2.CreateThread(cross2(cur), boardID, "First", "Body")
	boards2.CreateThread(cross2(cur), boardID, "Second", "Body")
	boards2.CreateThread(cross2(cur), boardID, "Third", "Body")

	// Get readonly board
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	board, _ = hub.GetBoard(cross, uint64(boardID))
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	threads := hub.GetThreads(cross, board.ID, 0, board.ThreadCount)

	for _, t := range threads {
		println(t.ID, t.Title)
	}
}

// Output:
// 1 First
// 2 Second
// 3 Third

// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_hub_3_b_filetest

// Test getting threads from a board
package z_hub_3_b_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

var board boards2.Board

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"))
	boardID := boards2.CreateBoard(cross(cur), "test123", false, false)

	// Create a couple of board threads
	boards2.CreateThread(cross(cur), boardID, "First", "Body")
	boards2.CreateThread(cross(cur), boardID, "Second", "Body")
	boards2.CreateThread(cross(cur), boardID, "Third", "Body")

	// Get readonly board
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	board, _ = boards2.GetBoard(uint64(boardID))
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	threads := boards2.GetThreads(board.ID, 0, board.ThreadCount)

	for _, t := range threads {
		println(t.ID, t.Title)
	}
}

// Output:
// 1 First
// 2 Second
// 3 Third

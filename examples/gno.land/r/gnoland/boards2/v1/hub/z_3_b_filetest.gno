// Test getting threads from a board
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var board hub.Board

func init() {
	testing.SetRealm(testing.NewUserRealm("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"))
	boardID := boards2.CreateBoard(cross, "test123", false, false)

	// Create a couple of board threads
	boards2.CreateThread(cross, boardID, "First", "Body")
	boards2.CreateThread(cross, boardID, "Second", "Body")
	boards2.CreateThread(cross, boardID, "Third", "Body")

	// Get readonly board
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	board, _ = hub.GetBoard(uint64(boardID))
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	threads := hub.GetThreads(board.ID, 0, board.ThreadCount)

	for _, t := range threads {
		println(t.ID, t.Title)
	}
}

// Output:
// 1 First
// 2 Second
// 3 Third

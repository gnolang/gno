// Test getting threads from an empty board
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/hub"
)

var boardID boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"))
	boardID = boards2.CreateBoard(cross, "test123", false, false)
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))
	boards := hub.GetThreads(uint64(boardID), 0, boards2.BoardCount())
	println(len(boards))
}

// Output:
// 0

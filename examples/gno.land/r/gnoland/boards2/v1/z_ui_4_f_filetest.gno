// Render thread which has been flagged.
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	boardName         = "BoardName"
)

var threadID boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a readonly board and then add a thread
	boardID := boards2.CreateBoard(cross, boardName, false, false)
	threadID = boards2.CreateThread(cross, boardID, "Foo", "Body")

	// Flag the thread
	boards2.SetFlaggingThreshold(cross, boardID, 1)
	boards2.FlagThread(cross, boardID, threadID, "")
}

func main() {
	path := boardName + "/" + threadID.String()
	println(boards2.Render(path))
}

// Output:
// ⚠ Thread has been flagged as inappropriate

// Render thread which has been flagged.
package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	boardName = "BoardName"
)

var threadID boards2.PostID

func init() {
	testing.SetRealm(std.NewUserRealm(owner))

	// Create a readonly board and then add a thread
	boardID := cross(boards2.CreateBoard)(boardName, false)
	threadID = cross(boards2.CreateThread)(boardID, "Foo", "Body", false)

	// Flag the thread
	cross(boards2.SetFlaggingThreshold)(boardID, 1)
	cross(boards2.FlagThread)(boardID, threadID, "")
}

func main() {
	path := boardName + "/" + threadID.String()
	println(boards2.Render(path))
}

// Output:
// Thread with ID: 1 has been flagged as inappropriate

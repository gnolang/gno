// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_thread_05_filetest

// Render thread which has been flagged.
package z_ui_thread_05_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	boardName         = "BoardName"
)

var threadID boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a readonly board and then add a thread
	boardID := boards2.CreateBoard(cross2(cur), boardName, false, false)
	threadID = boards2.CreateThread(cross2(cur), boardID, "Foo", "Body")

	// Flag the thread
	boards2.SetFlaggingThreshold(cross2(cur), boardID, 1)
	boards2.FlagThread(cross2(cur), boardID, threadID, "Reason")
}

func main() {
	path := boardName + "/" + threadID.String()
	println(boards2.Render(path))
}

// Output:
// ⚠ Thread has been flagged as [inappropriate](/r/gnoland/boards2/v1:BoardName/1/flagging-reasons)

// Render default board view.
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

func init() {
	testing.SetRealm(std.NewUserRealm(owner))

	// Create a board and then add 3 threads
	boardID := cross(boards2.CreateBoard)(boardName, false)

	// Create thread "A" with a single comment
	threadID := cross(boards2.CreateThread)(boardID, "A", "Body", false)
	cross(boards2.CreateReply)(boardID, threadID, 0, "Body")

	// The other 2 threads are created without comments
	cross(boards2.CreateThread)(boardID, "B", "Body", false)
	threadID = cross(boards2.CreateThread)(boardID, "C", "Body", false)

	// Repost thread "C" into a different board
	dstBoardID := cross(boards2.CreateBoard)("Bar", false)
	cross(boards2.CreateRepost)(boardID, threadID, "Title", "Body", dstBoardID)
}

func main() {
	println(boards2.Render(boardName))
}

// Output:
// # BoardName
// Board created by g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq on 2009-02-13 11:31pm UTC, #1
//
// [Create Thread](/r/gnoland/boards2/v1$help&func=CreateThread&boardID=1&body=&title=) • [Manage Board](?menu=manageBoard)
//
//
//
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:BoardName?order=desc)
//
// **≡ [A](/r/gnoland/boards2/v1:BoardName/1)**
// Created by g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq on 2009-02-13 11:31pm UTC
// **1 replies • 0 reposts**
//
// **≡ [B](/r/gnoland/boards2/v1:BoardName/3)**
// Created by g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
//
// **≡ [C](/r/gnoland/boards2/v1:BoardName/4)**
// Created by g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq on 2009-02-13 11:31pm UTC
// **0 replies • 1 reposts**

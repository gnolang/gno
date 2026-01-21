// Render board sorting threads from newest to oldest.
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	boardName         = "TestBoard"
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board and then add 3 threads
	boardID := boards2.CreateBoard(cross, boardName, false, false)

	boards2.CreateThread(cross, boardID, "A", "Body")
	boards2.CreateThread(cross, boardID, "B", "Body")
	boards2.CreateThread(cross, boardID, "C", "Body")
}

func main() {
	path := boardName + "?order=desc"
	println(boards2.Render(path))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › TestBoard
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, #1
// ↳ [Create Thread](/r/gnoland/boards2/v1:TestBoard/create-thread) • [Request Invite](/r/gnoland/boards2/v1$help&func=RequestInvite&boardID=1) • [Manage Board](?menu=manageBoard)
//
// ---
// Sort by: [newest first](/r/gnoland/boards2/v1:TestBoard?order=asc)
//
// ###### [C](/r/gnoland/boards2/v1:TestBoard/3)
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
//
// ###### [B](/r/gnoland/boards2/v1:TestBoard/2)
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
//
// ###### [A](/r/gnoland/boards2/v1:TestBoard/1)
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**

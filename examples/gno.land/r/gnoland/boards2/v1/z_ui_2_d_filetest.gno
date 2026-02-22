// Render readonly board.
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

	// Create a readonly board and then add a thread
	boardID := boards2.CreateBoard(cross, boardName, false, false)
	boards2.CreateThread(cross, boardID, "A", "Body")
	boards2.FreezeBoard(cross, boardID)
}

func main() {
	println(boards2.Render(boardName))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › TestBoard
// > [!WARNING] Info
// > Creating new threads and commenting are disabled within this board
//
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, #1
// [List Members](/r/gnoland/boards2/v1:TestBoard/members) • [Unfreeze Board](/r/gnoland/boards2/v1$help&func=UnfreezeBoard&boardID=1&replyID=&threadID=)
//
//
//
// > [!NOTE] 2 THINGS YOU NEED TO GET STARTED
// > To post and comment, **retrieve your Gno key** with [Adena](https://adena.app) or [gnokey](https://docs.gno.land/users/interact-with-gnokey). You need **at least 3,000 GNOT** in your account (balance requirement, not a transfer).
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:TestBoard?order=desc)
//
// ###### [A](/r/gnoland/boards2/v1:TestBoard/1)
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) `owner` on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**

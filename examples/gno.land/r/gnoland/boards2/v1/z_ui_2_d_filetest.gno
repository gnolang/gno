// Render readonly board.
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

	// Create a readonly board and then add a thread
	boardID := cross(boards2.CreateBoard)(boardName, false)
	cross(boards2.CreateThread)(boardID, "A", "Body", false)
	cross(boards2.FreezeBoard)(boardID)
}

func main() {
	println(boards2.Render(boardName))
}

// Output:
// # BoardName
// Board created by g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq on 2009-02-13 11:31pm UTC, #1
// _**Starting new threads and commenting is disabled**_
//
// [List Members](/r/gnoland/boards2/v1:BoardName/members) • [Unfreeze Board](/r/gnoland/boards2/v1$help&func=UnfreezeBoard&boardID=1&replyID=&threadID=)
//
//
//
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:BoardName?order=desc)
//
// **≡ [A](/r/gnoland/boards2/v1:BoardName/1)**
// Created by g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**

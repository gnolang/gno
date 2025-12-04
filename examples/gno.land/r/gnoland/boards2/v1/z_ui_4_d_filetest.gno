// Render thread from a readonly board.
// Rendered thread action links should be limited to readonly actions.
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
	boardID := boards2.CreateBoard(cross, boardName, false)
	threadID = boards2.CreateThread(cross, boardID, "Foo", "Body")
	boards2.FreezeBoard(cross, boardID)
}

func main() {
	path := boardName + "/" + threadID.String()
	println(boards2.Render(path))
}

// Output:
// # Foo
//
// Body
//
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC
// [Flag](/r/gnoland/boards2/v1$help&func=FlagThread&boardID=1&reason=&threadID=1) • [Repost](/r/gnoland/boards2/v1$help&func=CreateRepost&boardID=1&body=&destinationBoardID=&threadID=1&title=)

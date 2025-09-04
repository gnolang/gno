// Render default board view when there are no threads.
package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	boardName = "BoardName"
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.CreateBoard(cross, boardName, false)
}

func main() {
	println(boards2.Render(boardName))
}

// Output:
// # BoardName
// Board created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, #1
//
// [Create Thread](/r/gnoland/boards2/v1$help&func=CreateThread&boardID=1&body=&title=) • [Manage Board](?menu=manageBoard)
//
// ---
// ### This board doesn't have any threads
// Do you want to [start a new conversation](/r/gnoland/boards2/v1$help&func=CreateThread&boardID=1&body=&title=) in this board ?

// Render default board view when there are no threads.
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

	boards2.CreateBoard(cross, boardName, false, false)
}

func main() {
	println(boards2.Render(boardName))
}

// Output:
// # [Boards](/r/gnoland/boards2/v1) › TestBoard
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, #1
// ↳ [Create Thread](/r/gnoland/boards2/v1:TestBoard/create-thread) • [Request Invite](/r/gnoland/boards2/v1$help&func=RequestInvite&boardID=1) • [Manage Board](?menu=manageBoard)
//
//
//
// > [!NOTE] 3 THINGS YOU NEED TO GET STARTED
// > To post and comment you need: a **Gno address** (create one with [Adena](https://adena.app) or [gnokey](https://docs.gno.land/users/interact-with-gnokey)), **GNOT** to pay for transactions (get some from the [faucet](https://faucet.gno.land/)), and a **registered username** ([register yours](/r/gnoland/users/v1)).
// ---
// ### This board doesn't have any threads
// Do you want to [start a new conversation](/r/gnoland/boards2/v1:TestBoard/create-thread) in this board ?

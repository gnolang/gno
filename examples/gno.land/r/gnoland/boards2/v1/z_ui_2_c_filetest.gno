// Render board view with the manage board menu expanded.
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	boardName         = "BoardName"
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board and a thread
	boardID := boards2.CreateBoard(cross, boardName, false, false)
	boards2.CreateThread(cross, boardID, "A", "Body")
}

func main() {
	path := boardName + "?menu=manageBoard"
	println(boards2.Render(path))
}

// Output:
// # BoardName
// Board created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, #1
//
// [Create Thread](/r/gnoland/boards2/v1$help&func=CreateThread&boardID=1&body=&title=) • [Request Invite](/r/gnoland/boards2/v1$help&func=RequestInvite&boardID=1) • **Manage Board**
// ↳[Invite Member](/r/gnoland/boards2/v1$help&func=InviteMember&boardID=1&role=&user=) • [List Invite Requests](/r/gnoland/boards2/v1:BoardName/invites) • [List Members](/r/gnoland/boards2/v1:BoardName/members) • [List Banned Users](/r/gnoland/boards2/v1:BoardName/banned-users) • [Freeze Board](/r/gnoland/boards2/v1$help&func=FreezeBoard&boardID=1)
//
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:BoardName?menu=manageBoard&order=desc)
//
// **≡ [A](/r/gnoland/boards2/v1:BoardName/1)**
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**

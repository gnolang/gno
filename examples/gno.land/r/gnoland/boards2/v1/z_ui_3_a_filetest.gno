// Render board members view.
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

	boards2.CreateBoard(cross, boardName, false, false)
}

func main() {
	path := boardName + "/members"
	println(boards2.Render(path))
}

// Output:
// # BoardName Members
// ### These are the board members
// | Member | Role | Actions |
// | --- | --- | --- |
// | [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) | owner | [remove](/r/gnoland/boards2/v1$help&func=RemoveMember&boardID=1&member=g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) • [change role](/r/gnoland/boards2/v1$help&func=ChangeMemberRole&boardID=1&member=g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq&role=) |

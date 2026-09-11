// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_ui_board_members_00_filetest

// Render board members view.
package z_ui_board_members_00_filetest

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx"
	boardName         = "BoardName"
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross(cur), boardName, false, false)
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
// | [g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx](/u/g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx) | owner | [remove](/r/gnoland/boards2/v1$help&func=RemoveMember&boardID=1&member=g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx) • [change role](/r/gnoland/boards2/v1$help&func=ChangeMemberRole&boardID=1&member=g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx&role=) |

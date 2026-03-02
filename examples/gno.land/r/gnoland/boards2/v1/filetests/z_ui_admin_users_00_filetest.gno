// Render realm admin users view.
package main

import (
	boards2 "gno.land/r/gnoland/boards2/v1"
)

func main() {
	println(boards2.Render("admin-users"))
}

// Output:
// # Admin Users
// ### These are the admin users of the realm
// | Member | Role | Actions |
// | --- | --- | --- |
// | [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) | owner | [remove](/r/gnoland/boards2/v1$help&func=RemoveMember&boardID=0&member=g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) • [change role](/r/gnoland/boards2/v1$help&func=ChangeMemberRole&boardID=0&member=g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq&role=) |

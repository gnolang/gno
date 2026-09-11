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
// | [g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx](/u/g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx) | owner | [remove](/r/gnoland/boards2/v1$help&func=RemoveMember&boardID=0&member=g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx) • [change role](/r/gnoland/boards2/v1$help&func=ChangeMemberRole&boardID=0&member=g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx&role=) |

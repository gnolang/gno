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
// | [g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf](/u/g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf) | owner | [remove](/r/gnoland/boards2/v1$help&func=RemoveMember&boardID=0&member=g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf) • [change role](/r/gnoland/boards2/v1$help&func=ChangeMemberRole&boardID=0&member=g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf&role=) |

// Render board members view.
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
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
// | [g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh](/u/g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) | owner | [remove](/r/gnoland/boards2/v1$help&func=RemoveMember&boardID=1&member=g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh) • [change role](/r/gnoland/boards2/v1$help&func=ChangeMemberRole&boardID=1&member=g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh&role=) |

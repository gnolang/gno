package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/permissions"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	perms boards.Permissions
	bid   boards.ID
)

func init() {
	// Create a new permissions instance without users
	perms = permissions.New()

	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "foobar", false, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetPermissions(cross, bid, perms)

	// Owner that setted new board permissions is not a member of the new permissions
	println(boards2.IsMember(bid, owner))
}

// Output:
// false

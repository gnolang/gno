package main

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	bid           = boards.ID(0) // Operate on realm instead of individual boards
)

var perms boards.Permissions

func init() {
	// Create a new permissions instance without users
	perms = permissions.New()
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.SetPermissions(cross, bid, perms)

	// Owner that setted new permissions is not a member of the new permissions
	println(boards2.IsMember(bid, owner))
}

// Output:
// false

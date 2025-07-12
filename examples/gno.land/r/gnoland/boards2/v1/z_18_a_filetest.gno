package main

import (
	"std"
	"testing"

	"gno.land/p/nt/commondao"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	bid   = boards2.BoardID(0)                                      // Operate on realm instead of individual boards
)

var perms boards2.Permissions

func init() {
	// Create a new permissions instance without users
	perms = boards2.NewBasicPermissions(commondao.New())
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.SetPermissions(cross, bid, perms)

	// Owner that setted new permissions is not a member of the new permissions
	println(boards2.IsMember(bid, owner))
}

// Output:
// false

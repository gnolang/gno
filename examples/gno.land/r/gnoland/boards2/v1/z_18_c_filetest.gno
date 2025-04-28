package main

import (
	"std"
	"testing"

	"gno.land/p/nt/commondao"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	perms boards2.Permissions
	bid   boards2.BoardID
)

func init() {
	// Create a new permissions instance without users
	perms = boards2.NewDefaultPermissions(commondao.New())

	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("foobar")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.SetPermissions)(bid, perms)

	// Owner that setted new board permissions is not a member of the new permissions
	println(boards2.IsMember(bid, owner))
}

// Output:
// false

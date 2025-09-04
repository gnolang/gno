package main

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/p/nt/commondao"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	perms boards2.Permissions
	bid   boards2.BoardID
)

func init() {
	// Create a new permissions instance without users
	perms = boards2.NewBasicPermissions(commondao.New())

	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "foobar", false)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.SetPermissions(cross, bid, perms)

	// Owner that setted new board permissions is not a member of the new permissions
	println(boards2.IsMember(bid, owner))
}

// Output:
// false

package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	// Lock the realm without locking realm members
	testing.SetRealm(std.NewUserRealm(owner))
	cross(boards2.LockRealm)(false)

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.LockRealm)(true)
	println(boards2.IsRealmLocked())
	println(boards2.AreRealmMembersLocked())
}

// Output:
// true
// true

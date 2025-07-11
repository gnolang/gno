package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.LockRealm(cross, false)

	println(boards2.IsRealmLocked())
	println(boards2.AreRealmMembersLocked())
}

// Output:
// true
// false

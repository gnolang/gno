package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

func init() {
	// Lock the realm without locking realm members
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.LockRealm(cross, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.LockRealm(cross, true)

	println(boards2.IsRealmLocked())
	println(boards2.AreRealmMembersLocked())
}

// Output:
// true
// true

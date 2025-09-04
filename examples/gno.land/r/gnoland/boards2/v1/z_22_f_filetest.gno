package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	// Lock the realm without locking realm members
	testing.SetRealm(runtime.NewUserRealm(owner))
	boards2.LockRealm(cross, false)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.LockRealm(cross, true)

	println(boards2.IsRealmLocked())
	println(boards2.AreRealmMembersLocked())
}

// Output:
// true
// true

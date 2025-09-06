package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	boards2.LockRealm(cross, true)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	// Should fail because realm is already locked
	boards2.LockRealm(cross, true)
}

// Error:
// realm and members are locked

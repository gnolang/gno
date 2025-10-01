package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.LockRealm(cross, false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Should fail because realm is already locked
	boards2.LockRealm(cross, false)
}

// Error:
// realm is locked

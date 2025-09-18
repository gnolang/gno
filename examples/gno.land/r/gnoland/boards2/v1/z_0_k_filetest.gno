package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	boards2.CreateBoard(cross, "TEST123", false)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	// Should fail because board name already exists with a different casing
	boards2.CreateBoard(cross, "test123", false)
}

// Error:
// board already exists

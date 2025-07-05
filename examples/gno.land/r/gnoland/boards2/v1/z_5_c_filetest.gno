package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.RemoveMember(cross, 0, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // Operate on realm DAO instead of individual boards
}

// Error:
// member not found

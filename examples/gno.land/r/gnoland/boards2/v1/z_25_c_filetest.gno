package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var bid boards2.BoardID

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.RequestInvite(cross, bid)
}

// Error:
// caller is already a member

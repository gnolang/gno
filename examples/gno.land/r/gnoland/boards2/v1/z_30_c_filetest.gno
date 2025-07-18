package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var bid boards2.BoardID

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
	boards2.Ban(cross, bid, user, boards2.BanDay, "Unpolite behavior")
}

func main() {
	testing.SetRealm(std.NewUserRealm("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"))

	// Try to unban without unbanning permissions
	boards2.Unban(cross, bid, user, "")
}

// Error:
// unauthorized

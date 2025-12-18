package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	user  address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
)

var bid boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	// Try to ban a user without banning permissions
	boards2.Ban(cross, bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.BanDay, "Reason")
}

// Error:
// unauthorized

package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = chain.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var bid boards2.BoardID

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(user))

	// Try to ban a user without banning permissions
	boards2.Ban(cross, bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.BanDay, "Reason")
}

// Error:
// unauthorized

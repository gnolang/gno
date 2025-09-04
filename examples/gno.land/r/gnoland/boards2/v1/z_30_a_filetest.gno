package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var bid boards2.BoardID

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)
	boards2.Ban(cross, bid, user, boards2.BanDay, "Unpolite behavior")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.Unban(cross, bid, user, "")

	println(boards2.IsBanned(bid, user))
}

// Output:
// false

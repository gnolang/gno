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

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	boards2.RequestInvite(cross, 404)
}

// Error:
// board does not exist with ID: 404

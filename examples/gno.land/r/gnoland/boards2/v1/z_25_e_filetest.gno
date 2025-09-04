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

func main() {
	testing.SetRealm(runtime.NewUserRealm(user))

	boards2.RequestInvite(cross, 404)
}

// Error:
// board does not exist with ID: 404

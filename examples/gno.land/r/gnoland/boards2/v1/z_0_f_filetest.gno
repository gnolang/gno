package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	_ "gno.land/r/gnoland/users/v1" // Import realm to pre-register default users
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.CreateBoard(cross, "gnoland", false)
}

// Error:
// board name is a user name registered to a different user

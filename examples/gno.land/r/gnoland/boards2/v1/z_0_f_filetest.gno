package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	susers "gno.land/r/sys/users"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

func init() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/users/v1"))
	susers.RegisterUser(cross, "gnoland", address("g1g3lsfxhvaqgdv4ccemwpnms4fv6t3aq3p5z6u7"))
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross, "gnoland", false, false)
}

// Error:
// board name is a user name registered to a different user

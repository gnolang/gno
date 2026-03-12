package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	uinit "gno.land/r/sys/users/init"
)

const owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"

func init() {
	uinit.RegisterUser(cross, "gnoland", address("g1g3lsfxhvaqgdv4ccemwpnms4fv6t3aq3p5z6u7"))
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateBoard(cross, "gnoland", false, false)
}

// Error:
// board name is a user name registered to a different user

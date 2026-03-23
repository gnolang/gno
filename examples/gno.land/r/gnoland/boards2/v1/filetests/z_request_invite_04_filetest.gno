package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	user  address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
)

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.RequestInvite(cross, 404)
}

// Error:
// board does not exist with ID: 404

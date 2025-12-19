package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/permissions"
)

const owner address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.InviteMember(cross, 0, "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn", permissions.RoleGuest)
}

// Error:
// unauthorized

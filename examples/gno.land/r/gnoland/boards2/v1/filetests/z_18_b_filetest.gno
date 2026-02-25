package main

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	user address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	bid          = boards.ID(0)                               // Operate on realm instead of individual boards
)

var perms boards.Permissions

func init() {
	// Create a new permissions instance
	perms = permissions.New()
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.SetPermissions(cross, bid, perms)
}

// Error:
// unauthorized

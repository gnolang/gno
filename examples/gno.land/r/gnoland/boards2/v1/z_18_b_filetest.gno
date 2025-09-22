package main

import (
	"testing"

	"gno.land/p/nt/commondao"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	user = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	bid  = boards2.BoardID(0)                                  // Operate on realm instead of individual boards
)

var perms boards2.Permissions

func init() {
	// Create a new permissions instance
	perms = boards2.NewBasicPermissions(commondao.New())
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	boards2.SetPermissions(cross, bid, perms)
}

// Error:
// unauthorized

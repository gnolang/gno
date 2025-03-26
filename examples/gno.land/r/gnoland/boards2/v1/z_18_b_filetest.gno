package main

import (
	"std"
	"testing"

	"gno.land/p/gnoland/commondao"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	user = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	bid  = boards2.BoardID(0)                                      // Operate on realm instead of individual boards
)

var perms boards2.Permissions

func init() {
	// Create a new permissions instance
	perms = boards2.NewDefaultPermissions(commondao.New())

	testing.SetOriginCaller(user)
}

func main() {
	boards2.SetPermissions(bid, perms)
}

// Error:
// unauthorized

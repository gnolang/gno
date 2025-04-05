package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2

func init() {
	testing.SetOriginCaller(owner)
}

func main() {
	boards2.ChangeMemberRole(0, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleGuest)
}

// Error:
// unauthorized

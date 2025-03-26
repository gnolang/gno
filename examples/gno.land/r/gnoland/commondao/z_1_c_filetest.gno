package main

import (
	"std"
	"testing"

	"gno.land/r/gnoland/commondao"
)

const (
	user    = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	invitee = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

func init() {
	// Call as a users which is not a Common DAO member
	testing.SetOriginCaller(user)
}

func main() {
	commondao.Invite(invitee)
}

// Error:
// unauthorized

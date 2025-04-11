package main

import (
	"std"
	"testing"

	pcommondao "gno.land/p/gnoland/commondao"

	"gno.land/r/gnoland/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init() {
	// Use a realm caller
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	// Invite a user to be able to start creating DAOs
	testing.SetOriginCaller(owner)
	commondao.Invite(user)

	// Create a new DAO which gives ownership to `commondao_test`
	testing.SetOriginCaller(user)
	daoID = commondao.New("Foo").ID()
}

func main() {
	commondao.Vote(daoID, 1, pcommondao.ChoiceYes, "")
}

// Error:
// caller is not a DAO member

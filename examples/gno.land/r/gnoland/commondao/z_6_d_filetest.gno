package main

import (
	"std"

	pcommondao "gno.land/p/gnoland/commondao"

	"gno.land/r/gnoland/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var dao *pcommondao.CommonDAO

func init() {
	// Use a realm caller
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	// Invite a user to be able to start creating DAOs
	std.TestSetOriginCaller(owner)
	commondao.Invite(user)

	// Create a new DAO which gives ownership to `commondao_test`
	std.TestSetOriginCaller(user)
	dao = commondao.New("Foo")

	// Configure DAO
	dao.Members().Add(user)
}

func main() {
	commondao.Execute(dao.ID(), 404)
}

// Error:
// proposal not found

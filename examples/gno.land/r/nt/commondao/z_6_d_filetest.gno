package main

import (
	"std"
	"testing"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var dao *pcommondao.CommonDAO

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	// Create a new DAO which gives ownership to `commondao_test`
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	dao = commondao.New("Foo")

	// Configure DAO
	dao.Members().Add(user)
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	cross(commondao.Execute)(dao.ID(), 404)
}

// Error:
// proposal not found

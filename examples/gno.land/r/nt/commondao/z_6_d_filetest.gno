package main

import (
	"chain"
	"chain/runtime"
	"testing"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = chain.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var dao *pcommondao.CommonDAO

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(runtime.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(runtime.NewUserRealm(user))
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/test"))
	dao = commondao.New(cross, "Foo")

	// Configure DAO
	dao.Members().Add(user)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(user))

	commondao.Execute(cross, dao.ID(), 404)
}

// Error:
// proposal not found

package main

import (
	"chain/runtime"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(runtime.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// Create a DAO
	testing.SetRealm(runtime.NewUserRealm(user))
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/test"))
	daoID = commondao.New(cross, "Foo").ID()
}

func main() {
	// Use a caller that is not the owner of the DAO
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/tests/vm"))

	commondao.Get(daoID)
}

// Error:
// unauthorized

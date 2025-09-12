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

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(runtime.NewUserRealm(user))
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/nt/commondao"))
	dao := commondao.New(cross, "Foo", commondao.AllowExecution(false)) // Disallow execution
	daoID = dao.ID()
}

func main() {
	commondao.Execute(cross, daoID, 0)
}

// Error:
// proposal execution is not enabled

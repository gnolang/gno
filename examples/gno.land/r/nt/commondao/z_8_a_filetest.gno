package main

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = chain.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var rootID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(runtime.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(runtime.NewUserRealm(user))

	// Create root DAO with subDAOs
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/test"))
	rootDAO := commondao.New(cross, "Root DAO")
	fooDAO := commondao.NewSubDAO(cross, "A", rootDAO.ID())
	barDAO := commondao.NewSubDAO(cross, "B", fooDAO.ID())
	commondao.NewSubDAO(cross, "C", barDAO.ID())
	commondao.NewSubDAO(cross, "D", fooDAO.ID())

	rootID = rootDAO.ID()
}

func main() {
	testing.SetRealm(runtime.NewCodeRealm("gno.land/r/test"))

	iter := commondao.NewIterator(rootID)

	println("Count =", iter.Count())

	for iter.Next() {
		if dao := iter.DAO(); dao != nil {
			println(dao.Name())
		}
	}
}

// Output:
// Count = 4
// A
// B
// C
// D

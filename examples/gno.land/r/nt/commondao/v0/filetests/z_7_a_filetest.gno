package main

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "Foo"
)

var parentID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(testing.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/nt/commondao/v0"))
	parentDAO := commondao.New(cross, "Parent DAO", commondao.AllowChildren(true))
	parentID = parentDAO.ID()
}

func main() {
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/nt/commondao/v0"))

	dao := commondao.NewSubDAO(cross, name, parentID)
	if dao == nil {
		panic("expected subDAO to be created")
	}

	println(dao.Name() == name)
	println(dao.Parent().ID() == parentID)

	// Check that SubDAO is added as a child to the parent DAO
	if v := dao.Parent().Children().Get(0); v != nil {
		if subDAO, ok := v.(*pcommondao.CommonDAO); ok {
			println(subDAO.ID() == dao.ID())
		}
	}
}

// Output:
// true
// true
// true

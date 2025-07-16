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
	name  = "Foo"
)

var parentID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	commondao.Invite(cross, user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(std.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(std.NewCodeRealm("gno.land/r/nt/commondao"))
	parentDAO := commondao.New("Parent DAO", commondao.AllowChildren(true))
	parentID = parentDAO.ID()
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/nt/commondao"))

	dao := commondao.NewSubDAO(name, parentID)
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

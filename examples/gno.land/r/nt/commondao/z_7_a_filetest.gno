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
	slug  = "foo"
)

var rootDAO *pcommondao.CommonDAO

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(std.NewUserRealm(user))

	// Create the root DAO
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	rootDAO = commondao.New("Parent DAO")
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

	dao := commondao.NewSubDAO(name, slug, rootDAO)

	if dao == nil {
		panic("expected subDAO to be created")
	}

	println(dao.Name() == name)
	println(dao.Slug() == slug)
	println(dao.Parent().ID() == rootDAO.ID())

	// Check that SubDAO is added as a child to the parent DAO
	if v := dao.Parent().Children().Get(0); v != nil {
		if subDAO, ok := v.(*pcommondao.CommonDAO); ok {
			println(subDAO.Slug() == slug)
		}
	}
}

// Output:
// true
// true
// true
// true

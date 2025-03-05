// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"

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
	std.TestSetOriginCaller(owner)
	commondao.Invite(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	std.TestSetOriginCaller(user)

	// Create the root DAO
	rootDAO = commondao.New("Parent DAO")
}

func main() {
	dao := commondao.NewSubDAO(name, slug, rootDAO)

	if dao == nil {
		panic("expected subDAO to be created")
	}

	println(dao.Name() == name)
	println(dao.Slug() == slug)
	println(dao.Parent().ID() == rootDAO.ID())
}

// Output:
// true
// true
// true

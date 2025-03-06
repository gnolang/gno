// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "Foo"
	slug1 = "foo"
	slug2 = "bar"
	path  = "foo/bar"
)

var rootID, parentID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	std.TestSetOriginCaller(owner)
	commondao.Invite(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	std.TestSetOriginCaller(user)

	// Create root DAO with a subDAO
	rootDAO := commondao.New("Root DAO")
	parentDAO := commondao.NewSubDAO("Parent DAO", slug1, rootDAO)
	commondao.NewSubDAO(name, slug2, parentDAO)

	rootID = rootDAO.ID()
	parentID = parentDAO.ID()
}

func main() {
	dao := commondao.GetSubDAO(rootID, path)

	if dao == nil {
		panic("expected subDAO to be found")
	}

	println(dao.Name() == name)
	println(dao.Slug() == slug2)
	println(dao.Parent().ID() == parentID)
}

// Output:
// true
// true
// true

// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	name  = "Foo"
	slug  = "foo"
	path  = "foo"
)

var rootID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	std.TestSetOrigCaller(owner)
	commondao.Invite(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	std.TestSetOrigCaller(user)

	// Create root DAO with a subDAO
	rootDAO := commondao.New("Root DAO")
	commondao.NewSubDAO(name, slug, rootDAO)

	rootID = rootDAO.ID()
}

func main() {
	dao := commondao.GetSubDAO(rootID, path)

	if dao == nil {
		panic("expected subDAO to be found")
	}

	println(dao.Name() == name)
	println(dao.Slug() == slug)
	println(dao.Parent().ID() == rootID)
}

// Output:
// true
// true
// true

// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	name  = "B"
)

var daoID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	std.TestSetOriginCaller(owner)
	commondao.Invite(user)

	// Create a couple of DAOs
	std.TestSetOriginCaller(user)
	commondao.New("A")
	daoID = commondao.New(name).ID()
	commondao.New("C")
}

func main() {
	dao := commondao.Get(daoID)

	if dao == nil {
		panic("expected DAO to be found")
	}

	println(dao.Name() == name)
	println(dao.ID() == daoID)
}

// Output:
// true
// true

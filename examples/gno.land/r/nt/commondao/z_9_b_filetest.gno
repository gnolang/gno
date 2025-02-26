// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var rootID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	std.TestSetOrigCaller(owner)
	commondao.Invite(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	std.TestSetOrigCaller(user)

	// Create root DAO with subDAOs
	rootDAO := commondao.New("Root DAO")
	fooDAO := commondao.NewSubDAO("A", "foo", rootDAO)
	barDAO := commondao.NewSubDAO("B", "bar", fooDAO)
	commondao.NewSubDAO("C", "baz", barDAO)
	commondao.NewSubDAO("D", "baz", fooDAO)

	rootID = rootDAO.ID()
}

func main() {
	iter := commondao.NewIterator(rootID, commondao.WithCount(2))

	println("Count =", iter.Count())

	for iter.Next() {
		println("")
		if dao := iter.DAO(); dao != nil {
			println(dao.Name())
			println(dao.Slug())
		}
		println(iter.Path())
	}
}

// Output:
// Count = 2
//
// A
// foo
// foo
//
// B
// bar
// foo/bar

// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"
	"testing"

	"gno.land/r/gnoland/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var rootID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetOriginCaller(owner)
	commondao.Invite(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetOriginCaller(user)

	// Create root DAO with subDAOs
	rootDAO := commondao.New("Root DAO")
	fooDAO := commondao.NewSubDAO("A", "foo", rootDAO)
	barDAO := commondao.NewSubDAO("B", "bar", fooDAO)
	commondao.NewSubDAO("C", "baz", barDAO)
	commondao.NewSubDAO("D", "baz", fooDAO)

	rootID = rootDAO.ID()
}

func main() {
	iter := commondao.NewIterator(rootID, commondao.WithOffset(1), commondao.WithCount(2))

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
// B
// bar
// foo/bar
//
// C
// baz
// foo/bar/baz

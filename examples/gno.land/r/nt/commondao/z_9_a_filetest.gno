// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var rootID uint64

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(std.NewUserRealm(owner))
	cross(commondao.Invite)(user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(std.NewUserRealm(user))

	// Create root DAO with subDAOs
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))
	rootDAO := commondao.New("Root DAO")
	fooDAO := commondao.NewSubDAO("A", rootDAO)
	barDAO := commondao.NewSubDAO("B", fooDAO)
	commondao.NewSubDAO("C", barDAO)
	commondao.NewSubDAO("D", fooDAO)

	rootID = rootDAO.ID()
}

func main() {
	crossing()

	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/commondao_test"))

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

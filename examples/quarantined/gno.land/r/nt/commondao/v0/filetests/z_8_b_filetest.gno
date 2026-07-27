// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_8_b_filetest

package z_8_b_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var rootID uint64

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// The origin must be the invited user where invitation
	// is removed after the first user call to create a DAO
	testing.SetRealm(testing.NewUserRealm(user))

	// Create root DAO with subDAOs
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	rootDAO := commondao.New(cross(cur), "Root DAO", "")
	fooDAO := commondao.NewSubDAO(cross(cur), "A", rootDAO, "")
	barDAO := commondao.NewSubDAO(cross(cur), "B", fooDAO, "")
	commondao.NewSubDAO(cross(cur), "C", barDAO, "")
	commondao.NewSubDAO(cross(cur), "D", fooDAO, "")

	rootID = rootDAO
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))

	iter := commondao.NewIterator(cross(cur), rootID, commondao.WithCount(2))

	println("Count =", iter.Count())

	for iter.Next() {
		println(iter.DAO().Name())
	}
}

// Output:
// Count = 2
// A
// B

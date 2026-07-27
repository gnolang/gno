// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_9_a_filetest

package z_9_a_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Council is {user, owner}
	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", string(owner))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Resign(cross(cur), daoID)

	council := commondao.GetView(daoID).Council()
	println(council.Size())
	println(council.Has(user))
	println(council.Has(owner))
}

// Output:
// 1
// false
// true

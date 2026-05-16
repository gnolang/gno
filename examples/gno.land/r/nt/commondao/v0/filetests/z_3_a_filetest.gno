// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_3_a_filetest

package z_3_a_filetest

import (
	"chain"
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner    = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user     = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newRealm = address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
)

var (
	daoID     uint64
	realmAddr = chain.PackageAddress("gno.land/r/test")
)

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross2(cur), user)

	// Create a new DAO which gives ownership to `test`
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))
	daoID = commondao.New(cross2(cur), "Foo").ID()
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test"))

	commondao.TransferOwnership(cross2(cur), daoID, newRealm)

	println(commondao.IsOwner(realmAddr, daoID))
	println(commondao.IsOwner(newRealm, daoID))
}

// Output:
// false
// true

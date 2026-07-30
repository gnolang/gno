// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_f_filetest

package z_20_f_filetest

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

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// No-op deregister: the execution kind is opt-in and not seeded, so a
	// council vote to deregister an absent kind is rejected at creation.
	commondao.CreateDeregisterKindProposal(cross(cur), daoID, "execution")
}

// Error:
// proposal kind is not enabled

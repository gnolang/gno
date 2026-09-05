// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_19_f_filetest

package z_19_f_filetest

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
	// Self-brick guard: deregistering the manage-kinds kind itself would
	// leave the DAO unable to ever manage its kinds, so it fails at creation
	testing.SetRealm(testing.NewUserRealm(user))
	commondao.CreateDeregisterKindProposal(cross(cur), daoID, "manage-kinds")
}

// Error:
// the manage-kinds kind cannot be deregistered

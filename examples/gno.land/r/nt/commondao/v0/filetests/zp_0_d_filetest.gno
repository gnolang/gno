// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_0_d_filetest

package zp_0_d_filetest

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
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// The direct user caller must be invited. The invitation is removed after
	// the caller creates a DAO for the first time.
	testing.SetRealm(testing.NewUserRealm(user))

	// Create root DAO with a subDAO
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")

}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"))

	commondao.CreateCouncilUpdateProposal(cross(cur), daoID, "", "")
}

// Error:
// caller is not a council member

// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_9_g_filetest

package z_9_g_filetest

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
	daoID = commondao.New(cross(cur), "Foo", "", commondao.AllowTextProposals(true))

	commondao.CreateTextProposal(cross(cur), daoID, "First", "Body", 0)

	// Tightening the cap below the current count blocks new proposals
	// without touching existing ones
	commondao.UpdateOptions(cross(cur), daoID, commondao.MaxActiveProposals(1))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateTextProposal(cross(cur), daoID, "Second", "Body", 0)
}

// Error:
// max number of active proposals reached

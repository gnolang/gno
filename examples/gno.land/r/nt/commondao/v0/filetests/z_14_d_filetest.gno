// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_14_d_filetest

package z_14_d_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	rootID uint64
	midID  uint64
	leafID uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	rootID = commondao.New(cross(cur), "Root", "Purpose", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Mid", "Purpose", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	midID = 3
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), midID, "Leaf", "Purpose", "", string(user))
		commondao.Vote(cross(cur), midID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), midID, pID)
	}
	leafID = 4

	// Dissolve the leaf's entire ancestor chain
	dID := commondao.CreateDissolutionProposal(cross(cur), midID, "")
	commondao.Vote(cross(cur), rootID, dID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, dID)

	dID = commondao.CreateDissolutionProposal(cross(cur), rootID, owner)
	commondao.Vote(cross(cur), rootID, dID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, dID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Even a full orphan may only UNFREEZE itself, never freeze itself:
	// the rescue branch is guarded on !frozen.
	commondao.CreateTreasuryFreezeProposal(cross(cur), leafID, leafID, true)
}

// Error:
// a DAO cannot target itself

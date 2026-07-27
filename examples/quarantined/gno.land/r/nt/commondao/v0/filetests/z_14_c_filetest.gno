// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_14_c_filetest

package z_14_c_filetest

import (
	"chain"
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

	testing.IssueCoins(commondao.GetView(leafID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 100)))

	// The root (grandparent) freezes the leaf, then only the middle DAO
	// is dissolved. The root remains a live proper ancestor.
	fID := commondao.CreateTreasuryFreezeProposal(cross(cur), rootID, leafID, true)
	commondao.Vote(cross(cur), rootID, fID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, fID)

	dID := commondao.CreateDissolutionProposal(cross(cur), midID, "")
	commondao.Vote(cross(cur), rootID, dID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, dID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Self-unfreeze is still forbidden: a live proper ancestor (the root)
	// remains, so the orphan-rescue path does not open. The rescue
	// requires ALL proper ancestors dissolved, not just the parent.
	commondao.CreateTreasuryFreezeProposal(cross(cur), leafID, leafID, false)
}

// Error:
// a DAO cannot target itself

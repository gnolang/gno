// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_12_f_filetest

package z_12_f_filetest

import (
	"chain"
	"chain/banker"
	"strings"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

// Treasury clawback passes by SIMPLE majority. A 5-member ancestor council
// splitting 3 YES / 2 NO clears a simple majority (2*3 > 5) but not a
// supermajority (3*3 < 2*5), so under the correct Simple threshold the
// clawback executes and empties the target; a mutated Super threshold would
// dismiss it, leaving the funds.
const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	moul  = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
	mF    = address("g1cyh9qm4y43w38gfk77qxczrhqj2asdr8l7uk54")
	mG    = address("g1x95kxqpwvqjf98qpkvsay7tppkh6zs9fr7wwda")
)

var (
	rootID uint64
	subID  uint64
	pID    uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	members := strings.Join([]string{owner.String(), moul.String(), mF.String(), mG.String()}, "\n")
	rootID = commondao.New(cross(cur), "Root", "Purpose", "", members)

	// Create a funded sub-DAO of root (sub creation itself is simple
	// majority: 3 of 5 YES passes).
	sp := commondao.CreateSubDAOProposal(cross(cur), rootID, "Sub", "Purpose", "", string(user))
	testing.SetRealm(testing.NewUserRealm(user))
	commondao.Vote(cross(cur), rootID, sp, pdao.ChoiceYes, "")
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Vote(cross(cur), rootID, sp, pdao.ChoiceYes, "")
	testing.SetRealm(testing.NewUserRealm(moul))
	commondao.Vote(cross(cur), rootID, sp, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, sp)
	subID = 3
	testing.IssueCoins(commondao.GetView(subID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 500)))

	// Clawback proposal, hosted in the ancestor (root), 3 YES / 2 NO.
	pID = commondao.CreateTreasuryClawbackProposal(cross(cur), rootID, subID)
	testing.SetRealm(testing.NewUserRealm(mF))
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceNo, "")
	testing.SetRealm(testing.NewUserRealm(mG))
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceNo, "")
	testing.SetRealm(testing.NewUserRealm(user))
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	testing.SetRealm(testing.NewUserRealm(moul))
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SkipHeights(121000)

	commondao.Execute(cross(cur), rootID, pID)

	p, _ := commondao.GetView(rootID).GetProposal(pID)
	println("status:", string(p.Status()))
	println("sub treasury:", banker.NewReadonlyBanker().GetCoins(commondao.GetView(subID).Address()).String())
}

// Output:
// status: executed
// sub treasury:

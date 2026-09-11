// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_11_h_filetest

package z_11_h_filetest

import (
	"chain"
	"chain/banker"
	"strings"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

// A treasury spend requires a Supermajority (>= 2/3), not a simple
// majority. A 5-member council splitting 3 YES / 2 NO clears a simple
// majority (2*3 > 5) but NOT a supermajority (3*3 < 2*5), so at the
// deadline the spend is dismissed and the treasury is untouched. This
// pins the spend threshold behaviourally: flipping it to SimpleMajority
// would let this same vote pass and spend the funds.
const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	moul  = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
	mF    = address("g1cyh9qm4y43w38gfk77qxczrhqj2asdr8l7uk54")
	mG    = address("g1x95kxqpwvqjf98qpkvsay7tppkh6zs9fr7wwda")
)

var (
	daoID uint64
	pID   uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// A 5-member council (caller + 4 additional members).
	testing.SetRealm(testing.NewUserRealm(user))
	members := strings.Join([]string{owner.String(), moul.String(), mF.String(), mG.String()}, "\n")
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", members)

	testing.IssueCoins(commondao.GetView(daoID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	pID = commondao.CreateTreasurySpendProposal(cross(cur), daoID, owner, "ugnot", 400)

	// 3 YES / 2 NO from distinct council members. NO votes first so a
	// (would-be) simple-majority pass under a mutated threshold lands on
	// the last vote, never on an already-decided proposal.
	testing.SetRealm(testing.NewUserRealm(mF))
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceNo, "")
	testing.SetRealm(testing.NewUserRealm(mG))
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceNo, "")
	testing.SetRealm(testing.NewUserRealm(user))
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	testing.SetRealm(testing.NewUserRealm(moul))
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))
	testing.SkipHeights(121000) // 5s blocks: pass the 7 day voting deadline

	commondao.Execute(cross(cur), daoID, pID)

	view := commondao.GetView(daoID)
	p, _ := view.GetProposal(pID)
	println("status:", string(p.Status()))
	println("treasury:", banker.NewReadonlyBanker().GetCoins(view.Address()).String())
}

// Output:
// status: dismissed
// treasury: 1000ugnot

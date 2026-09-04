// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_7_g_filetest

package z_7_g_filetest

import (
	"strings"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

// Sub-DAO creation passes by SIMPLE majority (spec :1504), not
// supermajority. A 5-member council splitting 3 YES / 2 NO clears a simple
// majority (2*3 > 5) but not a supermajority (3*3 < 2*5), so under the
// correct Simple threshold the sub-DAO is created; a mutated Super threshold
// would dismiss it. Mirror of z_11_h (spend = Super).
const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	moul  = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
	mF    = address("g1cyh9qm4y43w38gfk77qxczrhqj2asdr8l7uk54")
	mG    = address("g1x95kxqpwvqjf98qpkvsay7tppkh6zs9fr7wwda")
)

var (
	rootID uint64
	pID    uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	members := strings.Join([]string{owner.String(), moul.String(), mF.String(), mG.String()}, "\n")
	rootID = commondao.New(cross(cur), "Root", "Purpose", "", members)

	pID = commondao.CreateSubDAOProposal(cross(cur), rootID, "Sub", "Purpose", "", string(user))

	// 3 YES / 2 NO from distinct members; NO first so the simple-majority
	// pass lands on the last (3rd YES) vote, never on a decided proposal.
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
	testing.SkipHeights(121000) // pass the 7 day voting deadline

	commondao.Execute(cross(cur), rootID, pID)

	p, _ := commondao.GetView(rootID).GetProposal(pID)
	println("status:", string(p.Status()))
	println("children:", commondao.GetView(rootID).ChildrenCount())
}

// Output:
// status: executed
// children: 1

// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_19_d_filetest

package z_19_d_filetest

import (
	"strings"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

// Proposal kind toggles are decided by SUPERMAJORITY: with a 5-member
// council, 3 YES votes (a simple majority) settle nothing (3*3 < 2*5), so
// the proposal stays active and cannot be executed before its deadline. A
// mutant deciding the toggle by simple majority would pass it early here.
const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	moul  = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
	memD  = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	memE  = address("g1cyh9qm4y43w38gfk77qxczrhqj2asdr8l7uk54")
)

var (
	daoID uint64
	pID   uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	// Council is {user, owner, moul, memD, memE}
	testing.SetRealm(testing.NewUserRealm(user))
	members := strings.Join([]string{owner.String(), moul.String(), memD.String(), memE.String()}, "\n")
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", members)

	pID = commondao.CreateSetProposalKindProposal(cross(cur), daoID, "text", false)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	testing.SetRealm(testing.NewUserRealm(moul))
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
}

func main(cur realm) {
	// 3 of 5 YES: passed under a simple majority, undecided under the
	// required supermajority
	p, _ := commondao.GetView(daoID).GetProposal(pID)
	if p.Status() != pdao.StatusActive {
		panic("expected the proposal to remain active")
	}

	// An undecided proposal cannot be executed before its voting deadline
	testing.SetRealm(testing.NewUserRealm(user))
	commondao.Execute(cross(cur), daoID, pID)
}

// Error:
// voting deadline not met

// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_17_a_filetest

package z_17_a_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

const (
	owner  = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user   = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	rescue = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
)

var (
	rootID uint64
	subID  uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	rootID = commondao.New(cross(cur), "Root", "Purpose", "", "")

	// A sub-DAO whose only council member is `user`
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Sub", "Purpose", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	subID = 3
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The parent (root), as a proper ancestor, adds a new member to the
	// sub-DAO's council and removes the original — the spec's rescue path
	// (:1531-1532). Hosted and voted in the root at supermajority.
	pID := commondao.CreateAncestorCouncilUpdateProposal(cross(cur), rootID, subID, string(rescue), string(user))
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, pID)

	view := commondao.GetView(subID)
	council := view.Council()
	println("size:", council.Size())
	println("has rescue:", council.Has(rescue))
	println("has old:", council.Has(user))
}

// Output:
// size: 1
// has rescue: true
// has old: false

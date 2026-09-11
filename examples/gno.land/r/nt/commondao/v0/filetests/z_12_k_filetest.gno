// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_12_k_filetest

// Only a council member may propose an ancestor clawback.
//
// The ancestry check authorizes which treasury may be swept; it says
// nothing about who may ask. Without this gate a stranger could fill a
// DAO's active-proposal cap with clawback items it never chose to
// consider.
package z_12_k_filetest

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx, genesis council only
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1, owns the tree below
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

	pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Sub", "Purpose", "", string(user))
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, pID)
	subID = 3
}

func main(cur realm) {
	// A valid ancestry (root over sub) proposed by a non-member.
	testing.SetRealm(testing.NewUserRealm(owner))

	commondao.CreateTreasuryClawbackProposal(cross(cur), rootID, subID)
}

// Error:
// caller is not a council member

// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_16_f_filetest

package z_16_f_filetest

import (
	"strconv"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
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
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")

	// Dissolve the root (root dissolution names a destination).
	pID := commondao.CreateDissolutionProposal(cross(cur), daoID, owner)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
}

func main() {
	// A dissolved root has no live ancestor, so its page carries both the
	// dissolved banner and the unrecoverable-deposit warning.
	println(commondao.Render(strconv.FormatUint(daoID, 10)))
}

// Output:
//
// > ⚠ This DAO has been dissolved
//
//
// > ⚠ No live ancestor DAO remains: coins sent to this treasury address can no longer be recovered.
//
// # Foo
// **Purpose:** Purpose
//
// [View Proposals](/r/nt/commondao/v0:2/proposals) • [View Settings](/r/nt/commondao/v0:2/settings)
//
// ---
// | Council |
// | --- |
// | g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 |
//
//
// ## Treasury
// - Address: g1vpah5sslxspfdvj9vk07rukek327dzmtxqkmk2
// - Balance: (empty)

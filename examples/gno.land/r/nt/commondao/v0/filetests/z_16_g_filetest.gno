// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_16_g_filetest

package z_16_g_filetest

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

var (
	rootID uint64
	leafID uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	rootID = commondao.New(cross(cur), "Root", "Purpose", "", "")

	// A sub-DAO of the (still-live) root.
	pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Leaf", "Purpose", "", string(user))
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, pID)
	leafID = 3

	// Dissolve the leaf (the proposal is hosted in its parent, the root).
	dID := commondao.CreateDissolutionProposal(cross(cur), leafID, "")
	commondao.Vote(cross(cur), rootID, dID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, dID)
}

func main() {
	// The leaf is dissolved but its root is alive, so the page shows the
	// dissolved banner WITHOUT the unrecoverable-treasury warning: a live
	// proper ancestor can still claw back the leaf's funds. This pins the
	// !hasLiveProperAncestor gate's negative branch.
	println(commondao.Render(strconv.FormatUint(leafID, 10)))
}

// Output:
//
// > ⚠ This DAO has been dissolved
//
// # Leaf
// **Purpose:** Purpose
//
// [View Proposals](/r/nt/commondao/v0:3/proposals) • [View Settings](/r/nt/commondao/v0:3/settings) • [Go to Parent DAO](/r/nt/commondao/v0:2)
//
// ---
// | Council |
// | --- |
// | g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 |
//
//
// ## Treasury
// - Address: g17pr3grpaag8ketn84dl2y0nncef6u3jw8ca9gl
// - Balance: (empty)

// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_10_d_filetest

package z_10_d_filetest

import (
	"chain"
	"strconv"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var subID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	rootID := commondao.New(cross(cur), "Root", "",
		commondao.AllowChildren(true),
		commondao.AllowTreasuryProposals(true),
	)
	subID = commondao.NewSubDAO(cross(cur), "Sub", rootID, "", commondao.AllowTreasuryProposals(true))

	testing.IssueCoins(commondao.GetView(subID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 1234)))

	fID := commondao.CreateTreasuryFreezeProposal(cross(cur), rootID, subID, true)
	commondao.Vote(cross(cur), rootID, fID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, fID)
}

func main() {
	// The DAO page shows the funded balance and the frozen warning; the
	// settings page shows the treasury proposals flag
	println(commondao.Render(strconv.FormatUint(subID, 10)))
	println(commondao.Render(strconv.FormatUint(subID, 10) + "/settings"))
}

// Output:
// # Sub
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
// - Balance: 1234ugnot
// **⚠ The treasury is frozen: spends cannot execute until a proper ancestor DAO unfreezes it**
//
// ## Create Proposal
// These are the proposal supported by this DAO:
//
// <gno-columns>
// [Treasury Spend](/r/nt/commondao/v0$help&func=CreateTreasurySpendProposal&amount=&daoID=3&denom=ugnot&to=)
//
// This type of proposal sends coins from the DAO's own treasury when it passes and is executed.
//
//
// <gno-columns-sep>
// [Treasury Clawback](/r/nt/commondao/v0$help&func=CreateTreasuryClawbackProposal&daoID=3&targetID=)
//
// This type of proposal sweeps a descendant DAO's treasury to the descendant's parent DAO.
//
//
// <gno-columns-sep>
// [Treasury Freeze](/r/nt/commondao/v0$help&func=CreateTreasuryFreezeProposal&daoID=3&frozen=true&targetID=)
//
// This type of proposal freezes or unfreezes a descendant DAO's treasury. While frozen, no treasury spend can execute.
//
//
// </gno-columns>
//
// # Sub: Settings
// [Go to DAO](/r/nt/commondao/v0:3)
//
// ---
// ## Options
// | Options | Values |
// | --- | --- |
// | Allow Render | true |
// | Allow SubDAOs | true |
// | Enable Voting | true |
// | Enable early execution | true |
// | Text Proposals | false |
// | Council update proposals | false |
// | SubDAO creation proposals | false |
// | Treasury proposals | true |
// | DAO dissolution proposals | false |

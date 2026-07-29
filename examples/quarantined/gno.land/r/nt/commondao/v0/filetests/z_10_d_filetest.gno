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
	rootID := commondao.New(cross(cur), "Root", "Purpose", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Sub", "Purpose", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	subID = 3

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
// - Balance: 1234ugnot
// **⚠ The treasury is frozen: spends cannot execute until a proper ancestor DAO unfreezes it**
//
// ## Create Proposal
// These are the proposal supported by this DAO:
//
// <gno-columns>
// [General Proposal](/r/nt/commondao/v0$help&func=CreateTextProposal&body=&daoID=3&title=&votingDays=7)
//
// This type of proposal is also known as text proposal which can be used for example to get consensus on initiatives without actually making any change on-chain.
//
//
// <gno-columns-sep>
// [Update Council](/r/nt/commondao/v0$help&func=CreateCouncilUpdateProposal&daoID=3&newMembers=&removeMembers=)
//
// This type of proposal can be used to add new council members to this DAO and also to remove existing ones.
//
// A single proposal allows new council members to be added and any number of existing ones removed within the same proposal.
//
//
// <gno-columns-sep>
// [Ancestor Council Update](/r/nt/commondao/v0$help&func=CreateAncestorCouncilUpdateProposal&daoID=3&newMembers=&removeMembers=&targetID=)
//
// This type of proposal lets this DAO, as an ancestor, add or remove council members of one of its descendant DAOs — the rescue path for a descendant whose council is stuck or empty.
//
//
// <gno-columns-sep>
// [Dissolve DAO](/r/nt/commondao/v0$help&func=CreateDissolutionProposal&daoID=3&destination=)
//
// This type of proposal can be used to dissolve DAOs and SubDAOs.
//
// Dissolving a DAO can't be undone, once the dissolution proposal passes and is executed DAO will be readonly.
//
//
// <gno-columns-sep>
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
// <gno-columns-sep>
// [New SubDAO](/r/nt/commondao/v0$help&func=CreateSubDAOProposal&daoID=3&description=&members=&name=&purpose=)
//
// This type of proposal is used to create SubDAOs, which are used to create tree based DAOs.
//
//
// <gno-columns-sep>
// [Set Proposal Kind](/r/nt/commondao/v0$help&func=CreateSetProposalKindProposal&daoID=3&enabled=true&kindName=)
//
// This type of proposal enables or disables one of the realm's proposal kinds for this DAO. Disabling a kind blocks new proposals of that kind while in-flight ones still vote and execute.
//
//
// </gno-columns>
//
// # Sub: Settings
// [Go to DAO](/r/nt/commondao/v0:3)
//
// ---
// ## Info
// | Setting | Value |
// | --- | --- |
// | Listed | false |
// | Max active proposals | 32 |
// | Proposal kinds | ancestor\-council\-update, council\-update, dissolve, proposal\-kinds, subdao, text, treasury\-clawback, treasury\-freeze, treasury\-spend |

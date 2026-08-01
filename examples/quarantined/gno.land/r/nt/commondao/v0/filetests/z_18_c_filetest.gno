// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_18_c_filetest

package z_18_c_filetest

import (
	"strconv"
	"testing"

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
	// The Purpose is user-controlled Charter text: markdown metacharacters
	// must render escaped, never as live markup.
	daoID = commondao.New(cross(cur), "Foo", "*Evil* [x](https://evil.example) # H1", "", "")
}

func main() {
	println(commondao.Render(strconv.FormatUint(daoID, 10)))
}

// Output:
// # Foo
// **Purpose:** \*Evil\* \[x\]\(https://evil\.example\) \# H1
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
// ## Create Proposal
// These are the proposal supported by this DAO:
//
// <gno-columns>
// [General Proposal](/r/nt/commondao/v0$help&func=CreateTextProposal&body=&daoID=2&title=&votingDays=7)
//
// This type of proposal is also known as text proposal which can be used for example to get consensus on initiatives without actually making any change on-chain.
//
//
// <gno-columns-sep>
// [Update Council](/r/nt/commondao/v0$help&func=CreateCouncilUpdateProposal&daoID=2&newMembers=&removeMembers=)
//
// This type of proposal can be used to add new council members to this DAO and also to remove existing ones.
//
// A single proposal allows new council members to be added and any number of existing ones removed within the same proposal.
//
//
// <gno-columns-sep>
// [Ancestor Council Update](/r/nt/commondao/v0$help&func=CreateAncestorCouncilUpdateProposal&daoID=2&newMembers=&removeMembers=&targetID=)
//
// This type of proposal lets this DAO, as an ancestor, add or remove council members of one of its descendant DAOs — the rescue path for a descendant whose council is stuck or empty.
//
//
// <gno-columns-sep>
// [Dissolve DAO](/r/nt/commondao/v0$help&func=CreateDissolutionProposal&daoID=2&destination=)
//
// This type of proposal can be used to dissolve DAOs and SubDAOs.
//
// Dissolving a DAO can't be undone, once the dissolution proposal passes and is executed DAO will be readonly.
//
//
// <gno-columns-sep>
// [Treasury Spend](/r/nt/commondao/v0$help&func=CreateTreasurySpendProposal&amount=&daoID=2&denom=ugnot&to=)
//
// This type of proposal sends coins from the DAO's own treasury when it passes and is executed.
//
//
// <gno-columns-sep>
// [Treasury Clawback](/r/nt/commondao/v0$help&func=CreateTreasuryClawbackProposal&daoID=2&targetID=)
//
// This type of proposal sweeps a descendant DAO's treasury to the descendant's parent DAO.
//
//
// <gno-columns-sep>
// [Treasury Freeze](/r/nt/commondao/v0$help&func=CreateTreasuryFreezeProposal&daoID=2&frozen=true&targetID=)
//
// This type of proposal freezes or unfreezes a descendant DAO's treasury. While frozen, no funds can leave the treasury.
//
//
// <gno-columns-sep>
// [New SubDAO](/r/nt/commondao/v0$help&func=CreateSubDAOProposal&daoID=2&description=&members=&name=&purpose=)
//
// This type of proposal is used to create SubDAOs, which are used to create tree based DAOs.
//
//
// <gno-columns-sep>
// [Register Proposal Kind](/r/nt/commondao/v0$help&func=CreateRegisterKindProposal&daoID=2&kindName=) • [Deregister Proposal Kind](/r/nt/commondao/v0$help&func=CreateDeregisterKindProposal&daoID=2&kindName=)
//
// This type of proposal registers one of the realm's catalog proposal kinds on this DAO by name, or deregisters a kind. Deregistering a kind blocks new proposals of that kind while in-flight ones still vote and execute. The manage-kinds kind itself cannot be deregistered.
//
//
// <gno-columns-sep>
// [Amend Bylaws](/r/nt/commondao/v0$help&func=CreateAmendBylawsProposal&daoID=2&payload=)
//
// This type of proposal adds, amends or removes one of the DAO's bylaws and mandates documents with a diff patch pinned to the current document text. Build the payload with the AmendBylawsPayload query function.
//
//
// </gno-columns>

package main

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
	"gno.land/r/gov/dao/bridge"
	govdao "gno.land/r/gov/dao/v2"
)

func init() {
	memberFn := func() []membstore.Member {
		return []membstore.Member{
			{
				Address:     std.Address("g123"),
				VotingPower: 10,
			},
			{
				Address:     std.Address("g456"),
				VotingPower: 10,
			},
			{
				Address:     std.Address("g789"),
				VotingPower: 10,
			},
		}
	}

	// Create a proposal
	title := "new govdao member addition"
	description := "add new members to the govdao"

	prop := dao.ProposalRequest{
		Title:       title,
		Description: description,
		Executor:    govdao.NewMemberPropExecutor(memberFn),
	}

	bridge.GovDAO().Propose(prop)
}

func main() {
	println("--")
	println(govdao.GetMembStore().Size())
	println("--")
	println(govdao.Render(""))
	println("--")
	println(govdao.Render("0"))
	println("--")
	govdao.VoteOnProposal(0, "YES")
	println("--")
	println(govdao.Render("0"))
	println("--")
	println(govdao.Render(""))
	println("--")
	govdao.ExecuteProposal(0)
	println("--")
	println(govdao.Render("0"))
	println("--")
	println(govdao.Render(""))
	println("--")
	println(govdao.GetMembStore().Size())
}

// Output:
// --
// 1
// --
// # GovDAO Proposals
//
// ## [Prop #0 - new govdao member addition](/r/gov/dao/v2:0)
//
// **Status: ACTIVE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
//
// --
// # Proposal #0 - new govdao member addition
//
// ## Description
//
// add new members to the govdao
//
// ## Proposal information
//
// **Status: ACTIVE**
//
// **Voting stats:**
// - YES 0 (0%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTES 10 (100%)
//
//
// **Threshold met: FALSE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
// ### Actions
//
// #### [[Vote YES](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=YES)] - [[Vote NO](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=NO)] - [[Vote ABSTAIN](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=ABSTAIN)]
//
//
// --
// --
// # Proposal #0 - new govdao member addition
//
// ## Description
//
// add new members to the govdao
//
// ## Proposal information
//
// **Status: ACCEPTED**
//
// **Voting stats:**
// - YES 10 (100%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTES 0 (0%)
//
//
// **Threshold met: TRUE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
// ### Actions
//
// The voting period for this proposal is over.
//
//
//
//
// --
// # GovDAO Proposals
//
// ## [Prop #0 - new govdao member addition](/r/gov/dao/v2:0)
//
// **Status: ACCEPTED**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
//
// --
// --
// # Proposal #0 - new govdao member addition
//
// ## Description
//
// add new members to the govdao
//
// ## Proposal information
//
// **Status: EXECUTION SUCCESSFUL**
//
// **Voting stats:**
// - YES 10 (25%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTES 30 (75%)
//
//
// **Threshold met: FALSE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
// ### Actions
//
// The voting period for this proposal is over.
//
//
//
//
// --
// # GovDAO Proposals
//
// ## [Prop #0 - new govdao member addition](/r/gov/dao/v2:0)
//
// **Status: EXECUTION SUCCESSFUL**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
//
// --
// 4

// Events:
// [
//   {
//     "type": "ProposalAdded",
//     "attrs": [
//       {
//         "key": "proposal-id",
//         "value": "0"
//       },
//       {
//         "key": "proposal-author",
//         "value": "g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"
//       }
//     ],
//     "pkg_path": "gno.land/r/gov/dao/v2",
//     "func": "EmitProposalAdded"
//   },
//   {
//     "type": "VoteAdded",
//     "attrs": [
//       {
//         "key": "proposal-id",
//         "value": "0"
//       },
//       {
//         "key": "author",
//         "value": "g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"
//       },
//       {
//         "key": "option",
//         "value": "YES"
//       }
//     ],
//     "pkg_path": "gno.land/r/gov/dao/v2",
//     "func": "EmitVoteAdded"
//   },
//   {
//     "type": "ProposalAccepted",
//     "attrs": [
//       {
//         "key": "proposal-id",
//         "value": "0"
//       }
//     ],
//     "pkg_path": "gno.land/r/gov/dao/v2",
//     "func": "EmitProposalAccepted"
//   },
//   {
//     "type": "ProposalExecuted",
//     "attrs": [
//       {
//         "key": "proposal-id",
//         "value": "0"
//       },
//       {
//         "key": "exec-status",
//         "value": "accepted"
//       }
//     ],
//     "pkg_path": "gno.land/r/gov/dao/v2",
//     "func": "ExecuteProposal"
//   }
// ]

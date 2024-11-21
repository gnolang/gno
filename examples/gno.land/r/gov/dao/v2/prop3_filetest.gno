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
	description := "add new members to the govdao"

	prop := dao.ProposalRequest{
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
// - [Proposal #0](/r/gov/dao/v2:0) - (**active**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// add new members to the govdao
//
// Status: active
//
// Voting stats: YES 0 (0%), NO 0 (0%), ABSTAIN 0 (0%), MISSING VOTE 10 (100%)
//
// Threshold met: false
//
//
// --
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// add new members to the govdao
//
// Status: accepted
//
// Voting stats: YES 10 (100%), NO 0 (0%), ABSTAIN 0 (0%), MISSING VOTE 0 (0%)
//
// Threshold met: true
//
//
// --
// - [Proposal #0](/r/gov/dao/v2:0) - (**accepted**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// add new members to the govdao
//
// Status: execution successful
//
// Voting stats: YES 10 (25%), NO 0 (0%), ABSTAIN 0 (0%), MISSING VOTE 30 (75%)
//
// Threshold met: false
//
//
// --
// - [Proposal #0](/r/gov/dao/v2:0) - (**execution successful**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
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

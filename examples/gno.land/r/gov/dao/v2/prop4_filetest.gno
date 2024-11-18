package main

import (
	"gno.land/p/demo/dao"
	"gno.land/r/gov/dao/bridge"
	govdaov2 "gno.land/r/gov/dao/v2"
	"gno.land/r/sys/params"
)

func init() {
	mExec := params.NewStringPropExecutor("prop1.string", "value1")
	title := "Setting prop1.string param"
	comment := "setting prop1.string param"
	prop := dao.ProposalRequest{
		Title:       title,
		Description: comment,
		Executor:    mExec,
	}
	id := bridge.GovDAO().Propose(prop)
	println("new prop", id)
}

func main() {
	println("--")
	println(govdaov2.Render(""))
	println("--")
	println(govdaov2.Render("0"))
	println("--")
	bridge.GovDAO().VoteOnProposal(0, "YES")
	println("--")
	println(govdaov2.Render("0"))
	println("--")
	bridge.GovDAO().ExecuteProposal(0)
	println("--")
	println(govdaov2.Render("0"))
}

// Output:
// new prop 0
// --
// - [Proposal #0](/r/gov/dao/v2:0) - (**active**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// setting prop1.string param
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
// setting prop1.string param
//
// Status: accepted
//
// Voting stats: YES 10 (100%), NO 0 (0%), ABSTAIN 0 (0%), MISSING VOTE 0 (0%)
//
// Threshold met: true
//
//
// --
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// setting prop1.string param
//
// Status: execution successful
//
// Voting stats: YES 10 (100%), NO 0 (0%), ABSTAIN 0 (0%), MISSING VOTE 0 (0%)
//
// Threshold met: true

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
//     "type": "set",
//     "attrs": [
//       {
//         "key": "k",
//         "value": "prop1.string"
//       }
//     ],
//     "pkg_path": "gno.land/r/sys/params",
//     "func": ""
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

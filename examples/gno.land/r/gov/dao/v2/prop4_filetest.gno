package main

import (
	"gno.land/p/demo/dao"
	"gno.land/r/gov/dao/bridge"
	govdaov2 "gno.land/r/gov/dao/v2"
	"gno.land/r/sys/params"
)

func init() {
	/*{ // prop0
		membersFn := func() []std.Address {
			return []std.Address{
				std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"),
			}
		}
		mExec := bridge.GovDAO().NewMemberPropExecutor(membersFn)
		comment := "adding someone to vote"
		id := bridge.GovDAO().Propose(comment, mExec)
		bridge.GovDAO().ExecuteProposal(id)
	}*/

	{ // prop0
		mExec := params.NewStringPropExecutor("prop1.string", "value1")
		comment := "setting prop1.string param"
		prop := dao.ProposalRequest{
			Description: comment,
			Executor:    mExec,
		}
		id := bridge.GovDAO().Propose(prop)
		println("new prop", id)
	}

	{ // prop1
		mExec := params.NewInt64PropExecutor("prop2.int64", 12345)
		comment := "setting prop2.int64 param"
		prop := dao.ProposalRequest{
			Description: comment,
			Executor:    mExec,
		}
		id := bridge.GovDAO().Propose(prop)
		println("new prop", id)
	}
}

func main() {
	println("--")
	println(govdaov2.Render(""))
	println("--")
	println(govdaov2.Render("1"))
	println("--")
	bridge.GovDAO().VoteOnProposal(1, "YES")
	println("--")
	println(govdaov2.Render("1"))
	println("--")
	bridge.GovDAO().ExecuteProposal(1)
	println("--")
	println(govdaov2.Render("1"))
}

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
//     "type": "ProposalAdded",
//     "attrs": [
//       {
//         "key": "proposal-id",
//         "value": "1"
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
//         "value": "1"
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
//         "value": "1"
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
//         "value": "prop2.int64"
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
//         "value": "1"
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

// Output:
// new prop 0
// new prop 1
// --
// - [Proposal #0](/r/gov/dao/v2:0) - (**active**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
// - [Proposal #1](/r/gov/dao/v2:1) - (**active**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// # Prop #1
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// setting prop2.int64 param
//
// Status: active
//
// Voting stats: YAY 0 (0%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 10 (100%)
//
// Threshold met: false
//
//
// --
// --
// # Prop #1
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// setting prop2.int64 param
//
// Status: accepted
//
// Voting stats: YAY 10 (100%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 0 (0%)
//
// Threshold met: true
//
//
// --
// --
// # Prop #1
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// setting prop2.int64 param
//
// Status: execution successful
//
// Voting stats: YAY 10 (100%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 0 (0%)
//
// Threshold met: true

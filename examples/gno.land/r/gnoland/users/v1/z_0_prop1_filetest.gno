package main

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/testutils"

	users "gno.land/r/gnoland/users/v1"
	_ "gno.land/r/gov/dao/init" // so that the govdao.GovDAO initializer is executed
	govdao "gno.land/r/gov/dao/v2"
	susers "gno.land/r/sys/users"
)

// Test updating a name via GovDAO

func init() {
	c := std.GetOrigCaller()
	alice := testutils.TestAddress("alice")

	// Register alice
	std.TestSetOrigCaller(alice)
	std.TestSetRealm(std.NewUserRealm(alice))
	users.Register("alice123")

	// Prop to change name
	std.TestSetOrigCaller(c)
	std.TestSetRealm(std.NewUserRealm(c))
	ex := users.ProposeNewName(alice, "alice_new123")

	// Create a proposal
	prop := dao.ProposalRequest{
		Title:       "Change alice's name!",
		Description: "",
		Executor:    ex,
	}

	govdao.GovDAO.Propose(prop)
}

func main() {
	println("--")
	println(govdao.Render(""))
	println("--")
	println(govdao.Render("0"))
	println("--")
	govdao.GovDAO.VoteOnProposal(0, "YES")
	println("--")
	println(govdao.Render("0"))
	println("--")
	govdao.GovDAO.ExecuteProposal(0)
	println("--")
	println(govdao.Render("0"))

	data, _ := susers.ResolveName("alice_new123")
	println(data.Addr())
}

// Output:
// --
// # GovDAO Proposals
//
// ## [Prop #0 - Change alice's name!](/r/gov/dao/v2:0)
//
// **Status: ACTIVE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
//
// --
// # Proposal #0 - Change alice's name!
//
// ## Description
//
// No description provided.
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
// # Proposal #0 - Change alice's name!
//
// ## Description
//
// No description provided.
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
// --
// --
// # Proposal #0 - Change alice's name!
//
// ## Description
//
// No description provided.
//
// ## Proposal information
//
// **Status: EXECUTION SUCCESSFUL**
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
// g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh

// Events:
// [
//   {
//     "type": "Registered",
//     "attrs": [
//       {
//         "key": "name",
//         "value": "alice123"
//       },
//       {
//         "key": "address",
//         "value": "g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh"
//       }
//     ],
//     "pkg_path": "gno.land/r/sys/users",
//     "func": "RegisterUser"
//   },
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
//     "type": "Updated",
//     "attrs": [
//       {
//         "key": "alias",
//         "value": "alice_new123"
//       },
//       {
//         "key": "address",
//         "value": "g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh"
//       }
//     ],
//     "pkg_path": "gno.land/r/sys/users",
//     "func": "UpdateName"
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

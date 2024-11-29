package main

import (
	"time"

	"gno.land/p/demo/dao"
	gnoblog "gno.land/r/gnoland/blog"
	govdao "gno.land/r/gov/dao/v2"
)

func init() {
	ex := gnoblog.NewPostExecutor(
		"hello-from-govdao",                             // slug
		"Hello from GovDAO!",                            // title
		"This post was published by a GovDAO proposal.", // body
		time.Now().Format(time.RFC3339),                 // publication date
		"moul",                                          // authors
		"govdao,example",                                // tags
	)

	// Create a proposal
	title := "govdao blog post title"
	description := "post a new blogpost about govdao"

	prop := dao.ProposalRequest{
		Title:       title,
		Description: description,
		Executor:    ex,
	}

	govdao.Propose(prop)
}

func main() {
	println("--")
	println(govdao.Render(""))
	println("--")
	println(govdao.Render("0"))
	println("--")
	govdao.VoteOnProposal(0, "YES")
	println("--")
	println(govdao.Render("0"))
	println("--")
	println(gnoblog.Render(""))
	println("--")
	govdao.ExecuteProposal(0)
	println("--")
	println(govdao.Render("0"))
	println("--")
	println(gnoblog.Render(""))
}

// Output:
// --
// # GovDAO Proposals
//
// ## [Prop #0 - govdao blog post title](/r/gov/dao/v2:0)
//
// **Status: ACTIVE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
//
// --
// # Proposal #0 - govdao blog post title
//
// ## Description
//
// post a new blogpost about govdao
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
// # Proposal #0 - govdao blog post title
//
// ## Description
//
// post a new blogpost about govdao
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
// # gno.land's blog
//
// No posts.
// --
// --
// # Proposal #0 - govdao blog post title
//
// ## Description
//
// post a new blogpost about govdao
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
//
//
// --
// # gno.land's blog
//
// <div class='columns-3'><div>
//
// ### [Hello from GovDAO!](/r/gnoland/blog:p/hello-from-govdao)
//  13 Feb 2009
// </div></div>

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

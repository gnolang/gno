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
	description := "post a new blogpost about govdao"

	prop := dao.ProposalRequest{
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
// ## [Prop #0](/r/gov/dao/v2:0)
//
// #### Status: ACTIVE
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
//
// --
// # Proposal #0
//
// ## Description
//
// post a new blogpost about govdao
//
// ## Proposal information
//
// #### Status: active
//
// #### Voting stats:
// - YES 0 (0%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTE 10 (100%)
//
//
// #### Threshold met: false
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// ## Actions
//
// [Vote YES](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=YES) - [Vote NO](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=NO) - [Vote ABSTAIN](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=ABSTAIN)
//
//
// --
// --
// # Proposal #0
//
// ## Description
//
// post a new blogpost about govdao
//
// ## Proposal information
//
// #### Status: accepted
//
// #### Voting stats:
// - YES 10 (100%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTE 0 (0%)
//
//
// #### Threshold met: true
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// ## Actions
//
// The voting period for this proposal is over.
//
//
//
//
// --
// # Gnoland's Blog
//
// No posts.
// --
// --
// # Proposal #0
//
// ## Description
//
// post a new blogpost about govdao
//
// ## Proposal information
//
// #### Status: execution successful
//
// #### Voting stats:
// - YES 10 (100%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTE 0 (0%)
//
//
// #### Threshold met: true
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// ## Actions
//
// The voting period for this proposal is over.
//
//
//
//
// --
// # Gnoland's Blog
//
// <div class='columns-3'><div>
//
// ### [Hello from GovDAO!](/r/gnoland/blog:p/hello-from-govdao)
//  13 Feb 2009
// </div></div>

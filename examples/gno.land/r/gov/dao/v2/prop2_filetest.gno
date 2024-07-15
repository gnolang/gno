package main

import (
	"time"

	gnoblog "gno.land/r/gnoland/blog"
	govdao "gno.land/r/gov/dao/v2"
)

func init() {
	ex := gnoblog.NewPostExecutor(
		"hello-from-govdao",                             // slug
		"Hello from GovDAO!",                            // title
		"This post was published by a GovDAO proposal.", // body
		time.Now().Format(time.RFC3339),                 // publidation date
		"moul",                                          // authors
		"govdao,example",                                // tags
	)

	// Create a proposal.
	comment := "post a new blogpost about govdao"
	govdao.Propose(comment, ex)
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
// - [Proposal #0](/r/gov/dao/v2:0) - (**active**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// post a new blogpost about govdao
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
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// post a new blogpost about govdao
//
// Status: accepted
//
// Voting stats: YAY 10 (100%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 0 (0%)
//
// Threshold met: true
//
//
// --
// # Gnoland's Blog
//
// No posts.
// --
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// post a new blogpost about govdao
//
// Status: execution successful
//
// Voting stats: YAY 10 (100%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 0 (0%)
//
// Threshold met: true
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

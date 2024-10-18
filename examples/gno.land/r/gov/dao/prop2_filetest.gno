package main

import (
	"std"
	"time"

	"gno.land/p/demo/context"
	"gno.land/p/gov/proposal"
	gnoblog "gno.land/r/gnoland/blog"
	govdao "gno.land/r/gov/dao"
)

func init() {
	membersFn := func() []std.Address {
		return []std.Address{
			std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"),
		}
	}

	mExec := govdao.NewPropExecutor(membersFn)

	title := "adding someone to vote"
	description := "adding someone to vote"

	id := govdao.Propose(title, description, mExec)

	govdao.ExecuteProposal(id)

	executor := proposal.NewCtxExecutor(func(ctx context.Context) error {
		gnoblog.DaoAddPost(
			ctx,
			"hello-from-govdao",  // slug
			"Hello from GovDAO!", // title
			"This post was published by a GovDAO proposal.", // body
			time.Now().Format(time.RFC3339),                 // publication date
			"moul",                                          // authors
			"govdao,example",                                // tags
		)
		return nil
	})

	// Create a proposal.
	// XXX: payment
	title = "post a new blogpost about govdao"
	description = "post a new blogpost about govdao"
	govdao.Propose(title, description, executor)
}

func main() {
	println("--")
	println(govdao.Render(""))
	println("--")
	println(govdao.Render("1"))
	println("--")
	govdao.VoteOnProposal(1, "YES")
	println("--")
	println(govdao.Render("1"))
	println("--")
	println(gnoblog.Render(""))
	println("--")
	govdao.ExecuteProposal(1)
	println("--")
	println(govdao.Render("1"))
	println("--")
	println(gnoblog.Render(""))
}

// Output:
// --
// # GovDAO Proposals
//
// ## [Prop #0 - adding someone to vote](/r/gov/dao:0)
//
// #### Status: Succeeded
//
// ---
//
// ## [Prop #1 - post a new blogpost about govdao](/r/gov/dao:1)
//
// #### Status: Active
//
// ---
//
//
// --
// # Prop #1 - post a new blogpost about govdao
//
// ## Description
//
// post a new blogpost about govdao
//
// ## Proposal information
//
// #### Status: Active
//
// #### Votes:
//
// - YAY: 0
//
// - NAY: 0
//
// - Percent: 0
//
// - Members: 1
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// --
// --
// # Prop #1 - post a new blogpost about govdao
//
// ## Description
//
// post a new blogpost about govdao
//
// ## Proposal information
//
// #### Status: Accepted
//
// #### Votes:
//
// - YAY: 1
//
// - NAY: 0
//
// - Percent: 100
//
// - Members: 1
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// --
// # Gnoland's Blog
//
// No posts.
// --
// --
// # Prop #1 - post a new blogpost about govdao
//
// ## Description
//
// post a new blogpost about govdao
//
// ## Proposal information
//
// #### Status: Succeeded
//
// #### Votes:
//
// - YAY: 1
//
// - NAY: 0
//
// - Percent: 100
//
// - Members: 1
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// --
// # Gnoland's Blog
//
// <div class='columns-3'><div>
//
// ### [Hello from GovDAO!](/r/gnoland/blog:p/hello-from-govdao)
//  13 Feb 2009
// </div></div>

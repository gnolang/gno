package main

import (
	"std"
	"time"

	"gno.land/p/demo/context"
	"gno.land/p/gov/executor"
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

	comment := "adding someone to vote"

	id := govdao.Propose(comment, mExec)

	govdao.ExecuteProposal(id)

	ex := executor.NewCtxExecutor(func(ctx context.Context) error {
		gnoblog.DaoAddPost(
			ctx,
			"hello-from-govdao",  // slug
			"Hello from GovDAO!", // title
			"This post was published by a GovDAO proposal.", // body
			time.Now().Format(time.RFC3339),                 // publidation date
			"moul",                                          // authors
			"govdao,example",                                // tags
		)
		return nil
	})

	// Create a proposal.
	// XXX: payment
	comment = "post a new blogpost about govdao"
	govdao.Propose(comment, ex)
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
// - [0](/r/gov/dao:0) - adding someone to vote (**succeeded**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
// - [1](/r/gov/dao:1) - post a new blogpost about govdao (**active**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// # Prop #1
//
// post a new blogpost about govdao
//
// Status: active
//
// Voting status: YES: 0, NO: 0, percent: 0, members: 1
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
//
// --
// --
// # Prop #1
//
// post a new blogpost about govdao
//
// Status: accepted
//
// Voting status: YES: 1, NO: 0, percent: 100, members: 1
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
//
// --
// # Gnoland's Blog
//
// No posts.
// --
// --
// # Prop #1
//
// post a new blogpost about govdao
//
// Status: succeeded
//
// Voting status: YES: 1, NO: 0, percent: 100, members: 1
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
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

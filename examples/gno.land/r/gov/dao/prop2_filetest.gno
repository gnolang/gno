// PKGPATH: gno.land/r/foo/prop02
package prop02

import (
	"std"
	"time"

	"gno.land/p/gov/proposal"
	gnoblog "gno.land/r/gnoland/blog"
	govdao "gno.land/r/gov/dao"
)

func init() {
	executor := proposal.NewExecutor(func() error {
		gnoblog.ModAddPost(
			"hello-from-govdao",                             // slug
			"Hello from GovDAO!",                            // title
			"This post was published by a GovDAO proposal.", // body
			time.Now().Format(time.RFC3339),                 // publidation date
			"moul",                                          // authors
			"govdao,example",                                // tags
		)
		return nil
	})

	// Create a proposal.
	// XXX: payment
	comment := "post a new blogpost about govdao"
	govdao.Propose(comment, executor)
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
// - [/r/gov/dao:0](0) - post a new blogpost about govdao (by g1qzu77x7g9z0klhnavmewmhxcz4jcckflhl7sce)
// --
// # Prop#0
//
// post a new blogpost about govdao
// Status: active
// Author: g1qzu77x7g9z0klhnavmewmhxcz4jcckflhl7sce
// --
// --
// # Prop#0
//
// post a new blogpost about govdao
// Status: accepted
// Author: g1qzu77x7g9z0klhnavmewmhxcz4jcckflhl7sce
// --
// # Gnoland's Blog
//
// No posts.
// --
// @@@@@@@@@
// gno.land/r/foo/prop02
// #########
// --
// # Prop#0
//
// post a new blogpost about govdao
// Status: success
// Author: g1qzu77x7g9z0klhnavmewmhxcz4jcckflhl7sce
// --
// # Gnoland's Blog
//
// <div class='columns-3'><div>
//
// ### [Hello from GovDAO!](/r/gnoland/blog:p/hello-from-govdao)
//  13 Feb 2009
// </div></div>

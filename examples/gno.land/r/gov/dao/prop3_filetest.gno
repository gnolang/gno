package main

import (
	"std"

	"gno.land/r/gnoland/blog"
	govdao "gno.land/r/gov/dao"
	"gno.land/r/sys/params"
)

func init() {
	{ // prop0
		membersFn := func() []std.Address {
			return []std.Address{
				std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"),
			}
		}
		mExec := govdao.NewPropExecutor(membersFn)
		comment := "adding someone to vote"
		id := govdao.Propose(comment, mExec)
		govdao.ExecuteProposal(id)
	}

	{ // prop1
		mExec := params.NewStringPropExecutor("prop1.string", "value1")
		comment := "setting prop1.string param"
		id := govdao.Propose(comment, mExec)
	}
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
	govdao.ExecuteProposal(1)
	println("--")
	println(govdao.Render("1"))
}

// Events:

// Output:
// --
// - [0](/r/gov/dao:0) - adding someone to vote (**succeeded**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
// - [1](/r/gov/dao:1) - setting prop1.string param (**active**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// # Prop #1
//
// setting prop1.string param
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
// setting prop1.string param
//
// Status: accepted
//
// Voting status: YES: 1, NO: 0, percent: 100, members: 1
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
//
// --
// --
// # Prop #1
//
// setting prop1.string param
//
// Status: succeeded
//
// Voting status: YES: 1, NO: 0, percent: 100, members: 1
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm

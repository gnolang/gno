package main

import (
	"std"

	govdao "gno.land/r/gov/dao"
	"gno.land/r/sys/validators"
)

func init() {
	// Create the validators change proposal.
	changesFn := func() []validators.Change {
		return []validators.Change{
			// add a new validator.
			{Address: std.Address("g12345678"), Power: 1},
			// remove an existing validator.
			{Address: std.Address("g000000000"), Power: 0},
		}
	}

	// Wraps changesFn to emit a certified event only if executed from a
	// complete governance proposal process.
	executor := validators.NewPropExecutor(changesFn)

	// Create a proposal.
	// XXX: payment
	comment := "manual valset changes proposal example"
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
	println(validators.Render(""))
	println("--")
	govdao.ExecuteProposal(1)
	println("--")
	println(govdao.Render("1"))
	println("--")
	println(validators.Render(""))
}

// Output:
// --
// - [/r/gov/dao:0](0) - manual valset changes proposal example (by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
// --
// # Prop#0
//
// manual valset changes proposal example
// Status: active
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// --
// --
// # Prop#0
//
// manual valset changes proposal example
// Status: accepted
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// --
// No valset changes to apply.
// --
// --
// # Prop#0
//
// manual valset changes proposal example
// Status: success
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
// --
// Valset changes to apply:
// - g12345678 (1)
// - g000000000 (0)

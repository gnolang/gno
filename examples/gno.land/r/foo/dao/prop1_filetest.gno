// Please note that this package is intended for demonstration purposes only.
// You could execute this code (the init part) by running a `maketx run` command
// or by uploading a similar package to a personal namespace.
//
// For the specific case of validators, a `r/gnoland/valopers` will be used to
// organize the lifecycle of validators (register, etc), and this more complex
// contract will be responsible to generate proposals.
package main

import (
	"std"

	foodao "gno.land/r/foo/dao"
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
	foodao.Propose(comment, executor)
}

func main() {
	println("--")
	println(foodao.Render(""))
	println("--")
	println(foodao.Render("1"))
	println("--")
	foodao.VoteOnProposal(1, "YES")
	println("--")
	println(foodao.Render("1"))
	println("--")
	println(validators.Render(""))
	println("--")
	foodao.ExecuteProposal(1)
	println("--")
	println(foodao.Render("1"))
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

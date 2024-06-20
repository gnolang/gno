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

	"gno.land/p/sys/vals/types"
	govdao "gno.land/r/gov/dao"
	"gno.land/r/sys/vals"
)

func init() {
	// Create the validators change proposal.
	changesFn := func() []types.Validator {
		return []types.Validator{
			{
				Address:     std.Address("g12345678"),
				PubKey:      "pubkey",
				VotingPower: 10, // add a new validator
			},
			{
				Address:     std.Address("g000000000"),
				PubKey:      "pubkey",
				VotingPower: 10, // add a new validator
			},
			{
				Address:     std.Address("g000000000"),
				PubKey:      "pubkey",
				VotingPower: 0, // remove an existing validator
			},
		}
	}

	// Wraps changesFn to emit a certified event only if executed from a
	// complete governance proposal process.
	executor := vals.NewPropExecutor(changesFn)

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
	println(vals.Render(""))
	println("--")
	govdao.ExecuteProposal(1)
	println("--")
	println(govdao.Render("1"))
	println("--")
	println(vals.Render(""))
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
// - g000000000 (1)
// - g000000000 (0)

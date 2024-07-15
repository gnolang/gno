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

	"gno.land/p/gov/dao"
	pVals "gno.land/p/sys/validators"
	govdao "gno.land/r/gov/dao"
	"gno.land/r/sys/validators"
)

func init() {
	changesFn := func() []pVals.Validator {
		return []pVals.Validator{
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
	executor := validators.NewPropExecutor(changesFn)

	// Create a proposal.
	comment := "manual valset changes proposal example"
	govdao.Propose(comment, executor)
}

func main() {
	println("--")
	println(govdao.Render(""))
	println("--")
	println(govdao.Render("0"))
	println("--")
	govdao.VoteOnProposal(0, dao.YesVote)
	println("--")
	println(govdao.Render("0"))
	println("--")
	println(validators.Render(""))
	println("--")
	govdao.ExecuteProposal(0)
	println("--")
	println(govdao.Render("0"))
	println("--")
	println(validators.Render(""))
}

// Output:
// --
// - [Proposal #0](/r/gov/dao:0) - (**active**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// manual valset changes proposal example
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
// manual valset changes proposal example
//
// Status: accepted
//
// Voting stats: YAY 10 (100%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 0 (0%)
//
// Threshold met: true
//
//
// --
// No valset changes to apply.
// --
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// manual valset changes proposal example
//
// Status: execution successful
//
// Voting stats: YAY 10 (100%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 0 (0%)
//
// Threshold met: true
//
//
// --
// Valset changes:
// - #123: g12345678 (10)
// - #123: g000000000 (10)
// - #123: g000000000 (0)

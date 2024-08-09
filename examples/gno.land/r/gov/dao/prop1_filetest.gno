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

	pVals "gno.land/p/sys/validators"
	govdao "gno.land/r/gov/dao"
	"gno.land/r/sys/validators"
)

const daoPkgPath = "gno.land/r/gov/dao"

func init() {
	membersFn := func() []std.Address {
		return []std.Address{
			std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"),
		}
	}

	mExec := govdao.NewPropExecutor(membersFn)

	description := "adding someone to vote"
	title := "adding someone to vote"
	id := govdao.Propose(title, description, mExec)
	govdao.ExecuteProposal(id)

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
	// XXX: payment
	description = "manual valset changes proposal example"
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
// # GovDAO Proposals
//
// ## [Prop #0 - adding someone to vote](/r/gov/dao:0)
//
// #### Status: Succeeded
//
// ---
//
// ## [Prop #1 - adding someone to vote](/r/gov/dao:1)
//
// #### Status: Active
//
// ---
//
//
// --
// # Prop #1 - adding someone to vote
//
// ## Description
//
// manual valset changes proposal example
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
// # Prop #1 - adding someone to vote
//
// ## Description
//
// manual valset changes proposal example
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
// No valset changes to apply.
// --
// --
// # Prop #1 - adding someone to vote
//
// ## Description
//
// manual valset changes proposal example
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
// Valset changes:
// - #123: g12345678 (10)
// - #123: g000000000 (10)
// - #123: g000000000 (0)

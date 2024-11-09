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

	"gno.land/p/demo/dao"
	pVals "gno.land/p/sys/validators"
	govdao "gno.land/r/gov/dao/v2"
	validators "gno.land/r/sys/validators/v2"
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

	// Create a proposal
	title := "Valset change"
	description := "manual valset changes proposal example"

	prop := dao.ProposalRequest{
		Title:       title,
		Description: description,
		Executor:    executor,
	}

	govdao.Propose(prop)
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
// # GovDAO Proposals
//
// ## [Prop #0](/r/gov/dao/v2:0)
//
// #### Status: ACTIVE
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
//
// --
// # Proposal #0 -
//
// ## Description
//
// manual valset changes proposal example
//
// ## Proposal information
//
// #### Status: active
//
// #### Voting stats:
// - YES 0 (0%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTE 10 (100%)
//
//
// #### Threshold met: false
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// ## Actions
//
// [Vote YES](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=YES) - [Vote NO](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=NO) - [Vote ABSTAIN](/r/gov/dao/v2$help&func=VoteOnProposal&id=0&option=ABSTAIN)
//
//
// --
// --
// # Proposal #0 -
//
// ## Description
//
// manual valset changes proposal example
//
// ## Proposal information
//
// #### Status: accepted
//
// #### Voting stats:
// - YES 10 (100%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTE 0 (0%)
//
//
// #### Threshold met: true
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// ## Actions
//
// The voting period for this proposal is over.
//
//
//
//
// --
// No valset changes to apply.
// --
// --
// # Proposal #0 -
//
// ## Description
//
// manual valset changes proposal example
//
// ## Proposal information
//
// #### Status: execution successful
//
// #### Voting stats:
// - YES 10 (100%)
// - NO 0 (0%)
// - ABSTAIN 0 (0%)
// - MISSING VOTE 0 (0%)
//
//
// #### Threshold met: true
//
// #### Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// ## Actions
//
// The voting period for this proposal is over.
//
//
//
//
// --
// Valset changes:
// - #123: g12345678 (10)
// - #123: g000000000 (10)
// - #123: g000000000 (0)

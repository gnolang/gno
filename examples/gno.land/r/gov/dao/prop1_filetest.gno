// PKGPATH: gno.land/r/foo/prop01
//
// Please note that this package is intended for demonstration purposes only.
// You could execute this code (the init part) by running a `maketx run` command
// or by uploading a similar package to a personal namespace.
//
// For the specific case of validators, a `r/gnoland/valopers` will be used to
// organize the lifecycle of validators (register, etc), and this more complex
// contract will be responsible to generate proposals.
package prop01

import (
	"std"

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
// - [/r/gov/dao:0](0) - manual valset changes proposal example (by g1acws3q8u0hjl9pndjy8rqz58q3x05dnd6d99qa)
// --
// # Prop#0
//
// manual valset changes proposal example
// Status: active
// Author: g1acws3q8u0hjl9pndjy8rqz58q3x05dnd6d99qa
// --
// --
// # Prop#0
//
// manual valset changes proposal example
// Status: accepted
// Author: g1acws3q8u0hjl9pndjy8rqz58q3x05dnd6d99qa
// --
// No valset changes to apply.
// --
// --
// # Prop#0
//
// manual valset changes proposal example
// Status: succeeded
// Author: g1acws3q8u0hjl9pndjy8rqz58q3x05dnd6d99qa
// --
// Valset changes to apply:
// - g12345678 (10)
// - g000000000 (10)
// - g000000000 (0)

// Package prop1 is an example of proposal creation using a contract. In
// practice, it could be done with a `maketx run` with a similar code.
//
// For the specific case of validators, a r/gnoland/validators will be used to
// organize the lifecycle of validators (register, etc), and this more complex
// contract will be responsible to generate proposals.
package prop1

import (
	"std"

	govdao "gno.land/r/gov/dao"
	"gno.land/r/sys/validators"
)

func init() {
	// Create the validators change proposal.
	changesFn := func() []validators.Change {
		return []validators.Change{
			{Address: std.Address("g12345678"), Power: 1},  // add a new validator
			{Address: std.Address("g000000000"), Power: 0}, // remove an existing validator
		}
	}

	// Wraps changesFn to emit a certified event only if executed from a
	// complete governance proposal process.
	executor := validators.NewExecutor(changesFn)

	// Create a proposal.
	// XXX: payment
	govdao.Propose(govdao.Proposal{
		Comment:  "manual valset changes proposal example",
		Executor: executor,
	})
	done = true
}

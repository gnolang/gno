// package prop1 is an example of proposal creation using a contract. in practice it could be done with a `maketx run` with a similar code.
// for the specific case of validators, I believe that a r/gnoland/validators will be used to organize the lifecycle of validators (register, etc), and that this more complex contract will be responsible to generate proposals.
package prop1

import (
	"std"

	govdao "gno.land/r/gov/dao"
	"gno.land/r/sys/validators"
)

var done bool

func Propose() {
	if done {
		panic("already done")
	}

	// create the validators change proposal.
	changesFn := func() []validators.Change {
		return []validators.Change{
			{Address: std.Address("g12345678"), Power: 1},  // add a new validator
			{Address: std.Address("g000000000"), Power: 0}, // remove an existing validator
		}
	}

	// wraps changesFn to emit a certified event only if executed from a complete governance proposal process.
	executor := validators.NewExecutor(changesFn)

	// create a proposal
	// XXX: payment
	govdao.Propose(govdao.Proposal{
		Comment:  "manual valset changes proposal example",
		Executor: executor,
	})
	done = true
}

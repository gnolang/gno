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
	changesFn := func() []validators.Change {
		return []validators.Change{
			{Address: std.Address("g12345678"), Power: 1},  // add a new validator
			{Address: std.Address("g000000000"), Power: 0}, // remove an existing validator
		}
	}

	executor := validators.NewExecutor(changesFn)

	govdao.Propose(govdao.Proposal{
		Comment:  "manual valset changes",
		Executor: executor,
	})
	done = true
}

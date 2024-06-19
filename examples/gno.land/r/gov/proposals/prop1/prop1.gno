// Package prop1 is an example of proposal creation using a contract.
//
// Please note that this package is intended for demonstration purposes only.
// You could execute this code by running a `maketx run` command or by uploading
// a similar package to a personal namespace.
//
// For the specific case of validators, a `r/gnoland/valopers` will be used to
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
	executor := validators.NewPropExecutor(changesFn)

	// Create a proposal.
	// XXX: payment
	comment := "manual valset changes proposal example"
	govdao.Propose(comment, executor)
}

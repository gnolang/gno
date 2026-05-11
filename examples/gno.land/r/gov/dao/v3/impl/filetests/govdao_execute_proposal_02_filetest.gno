package main

import "gno.land/r/gov/dao/v3/impl"

var govdao = impl.NewGovDAO()

func main() {
	// Try to execute a proposal using a nil executor
	govdao.ExecuteProposal(0, nil)
}

// Error:
// an executor is required to execute the proposal

package main

import (
	"gno.land/p/demo/dao"
	govdao "gno.land/r/gov/dao/v2"
	"gno.land/r/sys/vars"
)

func init() {
	changesFn := func() []vars.KV {
		return []vars.KV{
			{
				Key:   "key 1",
				Value: "value 1",
			},
			{
				Key:   "key 2",
				Value: "value 2",
			},
		}
	}

	// Create the executor
	executor := vars.NewVarPropExecutor(changesFn)

	// Create the proposal
	description := "Example value setting"

	prop := dao.ProposalRequest{
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
	govdao.ExecuteProposal(0)
	println("--")
	println(govdao.Render("0"))
	println("--")
	println(vars.GetStringValue("key 1"))
	println("--")
	println(vars.GetStringValue("key 2"))
}

// Output:
// --
// - [Proposal #0](/r/gov/dao/v2:0) - (**active**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// Example value setting
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
// Example value setting
//
// Status: accepted
//
// Voting stats: YAY 10 (100%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 0 (0%)
//
// Threshold met: true
//
//
// --
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// Example value setting
//
// Status: execution successful
//
// Voting stats: YAY 10 (100%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 0 (0%)
//
// Threshold met: true
//
//
// --
// value 1
// --
// value 2

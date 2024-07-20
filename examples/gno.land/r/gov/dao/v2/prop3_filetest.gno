package main

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
	govdao "gno.land/r/gov/dao/v2"
)

func init() {
	memberFn := func() []membstore.Member {
		return []membstore.Member{
			{
				Address:     std.Address("g123"),
				VotingPower: 10,
			},
			{
				Address:     std.Address("g456"),
				VotingPower: 10,
			},
			{
				Address:     std.Address("g789"),
				VotingPower: 10,
			},
		}
	}

	// Create a proposal
	description := "add new members to the govdao"

	prop := dao.ProposalRequest{
		Description: description,
		Executor:    govdao.NewMemberPropExecutor(memberFn),
	}

	govdao.Propose(prop)
}

func main() {
	println("--")
	println(govdao.GetMembStore().Size())
	println("--")
	println(govdao.Render(""))
	println("--")
	println(govdao.Render("0"))
	println("--")
	govdao.VoteOnProposal(0, "YES")
	println("--")
	println(govdao.Render("0"))
	println("--")
	println(govdao.Render(""))
	println("--")
	govdao.ExecuteProposal(0)
	println("--")
	println(govdao.Render("0"))
	println("--")
	println(govdao.Render(""))
	println("--")
	println(govdao.GetMembStore().Size())
}

// Output:
// --
// 1
// --
// - [Proposal #0](/r/gov/dao/v2:0) - (**active**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// add new members to the govdao
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
// add new members to the govdao
//
// Status: accepted
//
// Voting stats: YAY 10 (100%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 0 (0%)
//
// Threshold met: true
//
//
// --
// - [Proposal #0](/r/gov/dao/v2:0) - (**accepted**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// --
// # Prop #0
//
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// add new members to the govdao
//
// Status: execution successful
//
// Voting stats: YAY 10 (0%), NAY 0 (0%), ABSTAIN 0 (0%), HAVEN'T VOTED 30 (0%)
//
// Threshold met: false
//
//
// --
// - [Proposal #0](/r/gov/dao/v2:0) - (**execution successful**)(by g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm)
//
// --
// 4

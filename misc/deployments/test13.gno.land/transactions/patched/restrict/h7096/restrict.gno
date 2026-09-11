package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/params"
)

func main(cur realm) {
	r := params.ProposeRemoveUnrestrictedAcctsRequest(cross(cur),
		address("g148583t5x66zs6p90ehad6l4qefeyaf54s69wql"),
	)
	pid := dao.MustCreateProposal(cross(cur), r)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	dao.ExecuteProposal(cross(cur), pid)
}

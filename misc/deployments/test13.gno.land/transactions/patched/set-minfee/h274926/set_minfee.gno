package main

import (
	"gno.land/r/gnops/valopers/proposal"
	"gno.land/r/gov/dao"
)

func main(cur realm) {
	r := proposal.ProposeNewMinFeeProposalRequest(cross(cur), int64(0))
	pid := dao.MustCreateProposal(cross(cur), r)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	dao.ExecuteProposal(cross(cur), pid)
}

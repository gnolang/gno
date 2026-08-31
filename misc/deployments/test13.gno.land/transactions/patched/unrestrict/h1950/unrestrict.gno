package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/params"
)

func main(cur realm) {
	r := params.ProposeAddUnrestrictedAcctsRequest(cross(cur),
		address("g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa"),
	)
	pid := dao.MustCreateProposal(cross(cur), r)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	dao.ExecuteProposal(cross(cur), pid)
}

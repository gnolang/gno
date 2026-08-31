package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/cla"
)

func main(cur realm) {
	r := cla.ProposeNewCLA(cross(cur), "de2f507e38e514eca3329f8515435e3418315c10d81f62767ac9bf8cd7c78fad", "https://raw.githubusercontent.com/gnolang/gno/6022945545d58801f8740dc9db259818409f9275/CLA.md")
	pid := dao.MustCreateProposal(cross(cur), r)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	dao.ExecuteProposal(cross(cur), pid)
}

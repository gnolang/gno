package main

import (
	"std"
	"time"

	"gno.land/p/demo/dao"

	"gno.land/r/gnoland/users"
	_ "gno.land/r/gov/dao/init" // so that the govdao.GovDAO initializer is executed
	govdao "gno.land/r/gov/dao/v2"
)

func init() {
	ex := users.ProposeNewRelease("gno.land/r/gnoland/users/v2", "This is a note!")

	// Create a proposal
	prop := dao.ProposalRequest{
		Title:       "Propose users registry v2",
		Description: "",
		Executor:    ex,
	}

	govdao.GovDAO.Propose(prop)
}

func main() {
	govdao.GovDAO.VoteOnProposal(0, "YES")
	govdao.GovDAO.ExecuteProposal(0)
	println(users.Render(""))
}

// Output:
// # r/gnoland/users
// See the r/gnoland/users changelog below.
//
// | Version | Link | Notes |
// | --- | --- | --- |
// | v2 | [r/gnoland/users v2 (latest)](/r/gnoland/users/v2) | This is a note! |
// | v1 | [r/gnoland/users v1](/r/gnoland/users/v1) | [Original PR](https://github.com/gnolang/gno/pull/3166) |
//

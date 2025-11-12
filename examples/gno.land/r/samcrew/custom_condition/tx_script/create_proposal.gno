package main

import (
	"gno.land/p/samcrew/basedao"
	"gno.land/p/samcrew/daokit"
	"gno.land/r/samcrew/daodemo/custom_condition"
)

// Proposal function to add a new member in the DAO
// To execute this function, you must use a MsgRun `gnokey maketx run“
// See why it is necessary in Gno Documentation: https://docs.gno.land/users/interact-with-gnokey#run
func main() {
	payload := basedao.ActionAddMember{
		Address: "g1demo1234567890abcdefghijklmnopqrstuvwxyz",
		Roles:   []string{},
	}
	action := basedao.NewAddMemberAction(&payload)

	proposal := daokit.ProposalRequest{
		Title:       "Adding new member",
		Description: "Adding new member with no role",
		Action:      action,
	}

	custom_condition.Propose(cross, proposal)
}

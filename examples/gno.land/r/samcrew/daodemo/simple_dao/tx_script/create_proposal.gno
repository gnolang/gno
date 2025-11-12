package main

import (
	"gno.land/p/samcrew/basedao"
	"gno.land/p/samcrew/daokit"
	"gno.land/r/samcrew/daodemo/simple_dao"
)

// Proposal function to add a new member in the DAO
// To execute this function, you must use a MsgRun `gnokey maketx run“
// See why it is necessary in Gno Documentation: https://docs.gno.land/users/interact-with-gnokey#run
func main() {
	// Create payload for adding a new member with finance officer role
	payload := basedao.ActionAddMember{
		Address: "g1nqsjex9aguqy062wdrryvvg9u6dn6e9u44dmkm",
		Roles:   []string{"finance-officer"},
	}

	// Wrap payload in a DAO action structure
	action := basedao.NewAddMemberAction(&payload)

	// Create proposal request with metadata and action
	proposal := daokit.ProposalRequest{
		Title:       "Adding new member",
		Description: "Adding new finance-officer member.",
		Action:      action,
	}

	println("Proposal Title:", proposal.Title)
	println("Proposal Description:", proposal.Description)
	println("Action Type:", proposal.Action.Type())
	println("Action String:", proposal.Action.String())

	// Submit the proposal to our DAO
	simple_dao.Propose(cross, proposal)

}

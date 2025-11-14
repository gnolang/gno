package main

import (
	"gno.land/p/samcrew/daokit"
	"gno.land/r/samcrew/daodemo/custom_resource"
)

// Proposal function to add a new post on the blog
// (Displayed at the bottom of the main page)
// To execute this function, you must use a MsgRun `gnokey maketx run“
// See why it is necessary in Gno Documentation: https://docs.gno.land/users/interact-with-gnokey#run
func main() {
	// Calling the custom action
	action := custom_resource.NewPostAction("hello", "world")

	proposal := daokit.ProposalRequest{
		Title:       "Sample Proposal",
		Description: "This is a sample proposal using PrintAction.",
		Action:      action,
	}
	custom_resource.Propose(cross, proposal)
}

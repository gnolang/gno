package main

import(
	"time"
	"gno.land/r/volos/gov/governance"
)

// callback updates the minimum tokens required to create a proposal
func callback() {
	governance.SetProposalThreshold(1000000) // 1M tokens minimum
}

func main(){
	governance.CreateProposal(
		"Update Proposal Threshold", 
		"Increase minimum tokens required to create proposals to 1M tokens",
		"Ensures only serious proposals with sufficient community backing are submitted",
		time.Duration(7*24*time.Hour), // 7 days voting period
		callback,
	)
}
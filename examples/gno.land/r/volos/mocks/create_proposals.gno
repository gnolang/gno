package mocks

import (
	"time"
	"std"
	"gno.land/r/volos/gov/governance"
	"gno.land/r/volos/gov/vls"
	"gno.land/r/volos/gov/staker"
)

// callback function that will be executed if the proposal passes
func callback() {
	// This is a simple test callback that doesn't do anything
	// In a real scenario, this could update governance parameters, mint tokens, etc.
}

func CreateProposals(cur realm) {
	vls.Faucet(cross, 100000000)
	vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 100000000)
	staker.Stake(cross, 100000, std.DerivePkgAddr("gno.land/r/volos/mocks"))

	governance.CreateProposal(
		cross,
		"Test Proposal 1", 
		"This is a test proposal to verify governance functionality",
		time.Duration(1*time.Minute),
		callback,
	)

	governance.CreateProposal(
		cross,
		"Test Proposal 2", 
		"This is a test proposal to verify governance functionality",
		time.Duration(1*time.Minute),
		callback,
	)

	governance.CreateProposal(
		cross,
		"Test Proposal 3", 
		"This is a test proposal to verify governance functionality",
		time.Duration(5*time.Hour),
		callback,
	)
} 

package mocks

import (
	"std"
	"time"
	"gno.land/r/volos/gov/governance"
	"gno.land/r/volos/gov/vls"
	"gno.land/r/volos/gov/staker"
	"gno.land/r/volos/core"
)

// callback function that will be executed if the proposal passes
// This sets the fee to 25%
func setFeeCallback() {
	core.SetFee(cross, "gno.land/r/demo/wugnot:gno.land/r/gnoswap/v1/gns:3000:0", 25)
}

func setProposalThresholdCallback() {
	governance.SetProposalThreshold(50000)
}

func setVotingPowerQuorumCallback() {
	governance.SetVotingPowerQuorum(500)
}

func CreateProposals(cur realm) {
	vls.Faucet(cross, 100000000)
	vls.Approve(cross, std.DerivePkgAddr("gno.land/r/volos/gov/staker"), 100000000)
	staker.Stake(cross, 100000, std.DerivePkgAddr("gno.land/r/volos/mocks"))

	governance.CreateProposal(
		cross,
		"Set Fee to 15% - Short Duration", 
		"This proposal sets the market fee to 15% with a short voting period",
		time.Duration(1*time.Minute),
		setFeeCallback,
	)

	governance.CreateProposal(
		cross,
		"Set Proposal Threshold to 50000", 
		"This proposal sets the proposal threshold to 50000",
		time.Duration(1*time.Hour),
		setProposalThresholdCallback,
	)

	governance.CreateProposal(
		cross,
		"Set Voting Power Quorum to 500", 
		"This proposal sets the voting power quorum to 500",
		time.Duration(5*time.Hour),
		setVotingPowerQuorumCallback,
	)
} 

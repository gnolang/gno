// Package governance implements the Volos protocol's on-chain governance system.
//
// This package uses commondao as the core framework for proposals, voting,
// and execution logic. It is extended to support custom proposal types,
// voting power based on xVLS, and integration with the Volos DAO.
//
// Membership: To become a member, an address must hold at least MemberThreshold xVLS
// and call ApplyForMembership. Members are allowed to use their xVLS as voting power.
package governance

import (
	"std"
	"time"

	"gno.land/p/moul/authz"
	"gno.land/p/nt/commondao"
)

var (
	volosGovernance   *commondao.CommonDAO
	proposalThreshold int64 = 1000 //min xVLS required to create a proposal
	votingPowerQuorum int64 = 1000 //minimum xVLS voting power required for a proposal to pass (0 = disabled)
	maximumProposalDuration time.Duration = 14 * 24 * time.Hour //maximum duration of a proposal

	authorizer = authz.NewWithAuthority(authz.NewMemberAuthority(
		std.DerivePkgAddr("gno.land/r/volos/gov/staker"),
		std.DerivePkgAddr("gno.land/r/volos/gov/governance"),
	))
)

func init() {
	volosGovernance = commondao.New(
		commondao.WithName("Volos Governance"),
		commondao.WithDescription("On-chain governance for the Volos protocol."),
	)
}

// SetProposalThreshold allows governance to change the minimum xVLS required to propose.
func SetProposalThreshold(newThreshold int64) {
	authorizer.DoByCurrent("set_proposal_threshold", func() error {
		oldThreshold := proposalThreshold
		proposalThreshold = newThreshold
		caller := std.PreviousRealm().Address()
		emitProposalThresholdUpdated(caller, oldThreshold, newThreshold)
		return nil
	})
}

// SetVotingPowerQuorum allows governance to change the minimum xVLS voting power required for a proposal to pass.
func SetVotingPowerQuorum(newQuorum int64) {
	authorizer.DoByCurrent("set_voting_power_quorum", func() error {
		oldQuorum := votingPowerQuorum
		votingPowerQuorum = newQuorum
		caller := std.PreviousRealm().Address()
		emitVotingPowerQuorumUpdated(caller, oldQuorum, newQuorum)
		return nil
	})
}

// SetMaximumProposalDuration allows governance to change the maximum proposal duration.
func SetMaximumProposalDuration(newMax time.Duration) {
	authorizer.DoByCurrent("set_maximum_proposal_duration", func() error {
		maximumProposalDuration = newMax
		return nil
	})
}

// MemberSet returns a read-only view of the current member set for analytics and governance logic.
func MemberSet() commondao.MemberSet {
	return commondao.NewMemberSet(volosGovernance.Members())
}

// ProposalThreshold returns the current proposal threshold (for internal use).
func ProposalThreshold() int64 {
	return proposalThreshold
}

// VotingPowerQuorum returns the current voting power quorum requirement.
func VotingPowerQuorum() int64 {
	return votingPowerQuorum
}

// VotingPowerQuorumEnabled returns whether voting power quorum enforcement is enabled.
func VotingPowerQuorumEnabled() bool {
	return votingPowerQuorum > 0
}

// MaximumProposalDuration returns the current maximum allowed proposal duration.
func MaximumProposalDuration() time.Duration {
	return maximumProposalDuration
}

// Render function for explorer integration.
func Render(path string) string {
	return "Volos Governance - On-chain governance for the Volos protocol"
}

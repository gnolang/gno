package types

import (
	"std"
)

// Law defines all the business logic needed for the decision making of several features contained on the GovDAO.
type Law struct {
	// Tiers contains the actual valid tiers.
	Tiers TiersByName

	// VotingPercentageNeeded defines what is the value of positive votes needed to approve a proposal.
	VotingPercentageNeeded int

	// CheckMemberAddProposal is executed to check if the proposer is allowed to create a new proposal to add a new member.
	CheckMemberAddProposal func(membersByTier MembersByTier, proposer std.Address, proposedTier string, proposedAddr std.Address) error

	// CheckMemberPromoteProposal is executed to check if the proposer is allowed to promote the specified member.
	CheckMemberPromoteProposal func(membersByTier MembersByTier, proposer std.Address, proposedTier string, proposedAddr std.Address) error

	// CheckMemberWithdrawnProposal is executed to check if the proposer is allowed to withdraw the specific member.
	CheckMemberWithdrawnProposal func(membersByTier MembersByTier, proposer std.Address, proposedAddr std.Address) error

	// CheckProposalWithdrawn is executed to check if the proposer is allowed to withdraw a proposal.
	CheckProposalWithdrawn func(membersByTier MembersByTier, proposer std.Address) error
}
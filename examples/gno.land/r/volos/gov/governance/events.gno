package governance

import (
	"std"
	"strconv"
)

const (
	EventProposalCreated = "ProposalCreated"
	EventProposalExecuted = "ProposalExecuted"
	EventVoteCast        = "VoteCast"
	EventMemberAdded     = "MemberAdded"
	EventMemberRemoved   = "MemberRemoved"
	EventGovernanceUpdated = "GovernanceUpdated"
)

const (
	EventProposalIDKey   = "proposal_id"
	EventTitleKey        = "title"
	EventBodyKey         = "body"
	EventDeadlineKey     = "deadline"
	EventExecutedKey     = "executed"
	EventVoterKey        = "voter"
	EventVoteKey         = "vote"
	EventReasonKey       = "reason"
	EventMemberKey       = "member"
	EventOldPkgPathKey   = "old_pkg_path"
	EventNewPkgPathKey   = "new_pkg_path"
	EventCallerKey       = "caller"
	EventTimestampKey    = "timestamp"
	EventTotalMembersKey = "total_members"
	EventVoteCountKey    = "vote_count"
	EventYesVotesKey     = "yes_votes"
	EventNoVotesKey      = "no_votes"
	EventAbstainVotesKey = "abstain_votes"
)

// Emit proposal creation event
func emitProposalCreated(caller std.Address, proposalID uint64, title, body string, deadline int64) {
	std.Emit(
		EventProposalCreated,
		EventCallerKey, caller.String(),
		EventProposalIDKey, strconv.FormatUint(proposalID, 10),
		EventTitleKey, title,
		EventBodyKey, body,
		EventDeadlineKey, strconv.FormatInt(deadline, 10),
	)
}

// Emit proposal execution event
func emitProposalExecuted(caller std.Address, proposalID uint64) {
	std.Emit(
		EventProposalExecuted,
		EventCallerKey, caller.String(),
		EventProposalIDKey, strconv.FormatUint(proposalID, 10),
	)
}

// Emit vote cast event
func emitVoteCast(caller, voter std.Address, proposalID uint64, vote, reason string) {
	std.Emit(
		EventVoteCast,
		EventCallerKey, caller.String(),
		EventVoterKey, voter.String(),
		EventProposalIDKey, strconv.FormatUint(proposalID, 10),
		EventVoteKey, vote,
		EventReasonKey, reason,
	)
}

// Emit member added event
func emitMemberAdded(caller, member std.Address) {
	std.Emit(
		EventMemberAdded,
		EventCallerKey, caller.String(),
		EventMemberKey, member.String(),
		EventTotalMembersKey, strconv.FormatInt(int64(volosGovernance.Members().Size()), 10),
	)
}

// Emit member removed event
func emitMemberRemoved(caller, member std.Address) {
	std.Emit(
		EventMemberRemoved,
		EventCallerKey, caller.String(),
		EventMemberKey, member.String(),
		EventTotalMembersKey, strconv.FormatInt(int64(volosGovernance.Members().Size()), 10),
	)
}

// Emit governance updated event
func emitGovernanceUpdated(caller std.Address, oldPkgPath, newPkgPath string) {
	std.Emit(
		EventGovernanceUpdated,
		EventCallerKey, caller.String(),
		EventOldPkgPathKey, oldPkgPath,
		EventNewPkgPathKey, newPkgPath,
	)
} 

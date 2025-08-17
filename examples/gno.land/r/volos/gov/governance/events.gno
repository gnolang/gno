package governance

import (
	"std"
	"strconv"
	"time"
)

const (
	EventProposalCreated          = "ProposalCreated"
	EventProposalExecuted         = "ProposalExecuted"
	EventVoteCast                 = "VoteCast"
	EventMemberAdded              = "MemberAdded"
	EventMemberRemoved            = "MemberRemoved"
	EventGovernanceUpdated        = "GovernanceUpdated"
	EventVotingPowerQuorumUpdated = "VotingPowerQuorumUpdated"
	EventProposalThresholdUpdated = "ProposalThresholdUpdated"
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
	EventOldQuorumKey    = "old_quorum"
	EventNewQuorumKey    = "new_quorum"
	EventOldThresholdKey = "old_threshold"
	EventNewThresholdKey = "new_threshold"
	EventQuorumKey       = "quorum"
	EventXvlsAmountKey   = "xvls_amount"
)

// Emit proposal creation event
func emitProposalCreated(caller std.Address, proposalID uint64, title, body string, deadline, quorum int64) {
	std.Emit(
		EventProposalCreated,
		EventCallerKey, caller.String(),
		EventProposalIDKey, strconv.FormatUint(proposalID, 10),
		EventTitleKey, title,
		EventBodyKey, body,
		EventDeadlineKey, strconv.FormatInt(deadline, 10),
		EventQuorumKey, strconv.FormatInt(quorum, 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// Emit proposal execution event
func emitProposalExecuted(caller std.Address, proposalID uint64) {
	std.Emit(
		EventProposalExecuted,
		EventCallerKey, caller.String(),
		EventProposalIDKey, strconv.FormatUint(proposalID, 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// Emit vote cast event
func emitVoteCast(caller, voter std.Address, proposalID uint64, xvlsAmount int64, vote, reason string) {
	std.Emit(
		EventVoteCast,
		EventCallerKey, caller.String(),
		EventVoterKey, voter.String(),
		EventProposalIDKey, strconv.FormatUint(proposalID, 10),
		EventVoteKey, vote,
		EventReasonKey, reason,
		EventXvlsAmountKey, strconv.FormatInt(xvlsAmount, 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// Emit member added event
func emitMemberAdded(caller, member std.Address) {
	std.Emit(
		EventMemberAdded,
		EventCallerKey, caller.String(),
		EventMemberKey, member.String(),
		EventTotalMembersKey, strconv.FormatInt(int64(volosGovernance.Members().Size()), 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// Emit member removed event
func emitMemberRemoved(caller, member std.Address) {
	std.Emit(
		EventMemberRemoved,
		EventCallerKey, caller.String(),
		EventMemberKey, member.String(),
		EventTotalMembersKey, strconv.FormatInt(int64(volosGovernance.Members().Size()), 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// Emit governance updated event
func emitGovernanceUpdated(caller std.Address, oldPkgPath, newPkgPath string) {
	std.Emit(
		EventGovernanceUpdated,
		EventCallerKey, caller.String(),
		EventOldPkgPathKey, oldPkgPath,
		EventNewPkgPathKey, newPkgPath,
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// Emit voting power quorum updated event
func emitVotingPowerQuorumUpdated(caller std.Address, oldQuorum, newQuorum int64) {
	std.Emit(
		EventVotingPowerQuorumUpdated,
		EventCallerKey, caller.String(),
		EventOldQuorumKey, strconv.FormatInt(oldQuorum, 10),
		EventNewQuorumKey, strconv.FormatInt(newQuorum, 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// Emit proposal threshold updated event
func emitProposalThresholdUpdated(caller std.Address, oldThreshold, newThreshold int64) {
	std.Emit(
		EventProposalThresholdUpdated,
		EventCallerKey, caller.String(),
		EventOldThresholdKey, strconv.FormatInt(oldThreshold, 10),
		EventNewThresholdKey, strconv.FormatInt(newThreshold, 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

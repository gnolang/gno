package governance

import (
	"std"
	"time"

	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/demo/avl"

	en "gno.land/r/gnoswap/v1/emission"

	"gno.land/r/gnoswap/v1/common"

	"gno.land/p/demo/ufmt"
)

type voteWithWeight struct {
	Yes         bool
	Weight      uint64
	VotedHeight uint64
	VotedAt     uint64
}

var (
	votes         = avl.NewTree()           // voteKey(proposalId:user) -> yes/no
	userVotes     = avl.NewTree()           // user -> proposalId -> voteWithWeight
	accumYesVotes = make(map[uint64]uint64) // proposalId -> yes
	accumNoVotes  = make(map[uint64]uint64) // proposalId -> no
)

// createVoteKey creates a vote key for a given proposal ID and voter address
func createVoteKey(pid uint64, voter string) string {
	return ufmt.Sprintf("%d:%s", pid, voter)
}

// getVote gets a vote for a given proposal ID and voter address
// Returns:
// - vote: the vote for the given proposal ID and voter address
// - exists: a boolean indicating if user has voted on the proposal
func getVote(pid uint64, voter string) (bool, bool) {
	value, exists := votes.Get(createVoteKey(pid, voter))
	if !exists {
		return false, false
	}
	return value.(bool), true
}

func setVote(pid uint64, voter string, vote bool) {
	voteKey := createVoteKey(pid, voter)
	votes.Set(voteKey, vote)
}

// createUserVoteKey creates a user vote key for a given voter address and proposal ID
func createUserVoteKey(voter std.Address, pid uint64) string {
	return ufmt.Sprintf("%s:%d", voter.String(), pid)
}

func getUserVote(voter std.Address, pid uint64) (voteWithWeight, bool) {
	value, exists := userVotes.Get(createUserVoteKey(voter, pid))
	if !exists {
		return voteWithWeight{}, false
	}
	return value.(voteWithWeight), true
}

func setUserVote(voter std.Address, pid uint64, vote voteWithWeight) {
	if vote.Yes {
		accumYesVotes[pid] += vote.Weight
	} else {
		accumNoVotes[pid] += vote.Weight
	}
	userVotes.Set(createUserVoteKey(voter, pid), vote)
}

///////////////////// Vote /////////////////////

// Vote allows a user to vote on a given proposal.
// The user's voting weight is determined by their accumulated delegated stake until proposal creation time.
// ref: https://docs.gnoswap.io/contracts/governance/vote.gno#vote
func Vote(pid uint64, yes bool) string {
	common.IsHalted()
	en.MintAndDistributeGns()
	updateProposalsState()

	proposal := mustGetProposal(pid)

	voter := std.PrevRealm().Addr()
	now := uint64(time.Now().Unix())

	voteKey := createVoteKey(pid, voter.String())

	state, err := newVoteState(proposal, voteKey, voter, now)
	if err != nil {
		panic(err)
	}

	if err := state.validate(); err != nil {
		panic(err)
	}

	executor := newVoteExecutor(&proposal, voter, state.userWeight)
	if err := executor.execute(yes); err != nil {
		panic(err)
	}

	proposals.Set(formatUint(pid), proposal)

	setVote(pid, voter.String(), yes)
	setUserVote(voter, pid, voteWithWeight{
		Yes:         yes,
		Weight:      state.userWeight,
		VotedHeight: uint64(std.GetHeight()),
		VotedAt:     now,
	})

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"Vote",
		"prevAddr", prevAddr,
		"prevPkgPath", prevPkgPath,
		"proposalId", formatUint(pid),
		"voter", voter.String(),
		"yes", voteToString(yes),
		"voteWeight", formatUint(state.userWeight),
		"voteYes", formatUint(accumYesVotes[pid]),
		"voteNo", formatUint(accumNoVotes[pid]),
	)

	return voteKey
}

///////////////////// Vote State /////////////////////

type VoteState struct {
	isVotingPeriod  bool
	isProposalValid bool
	hasUserVoted    bool
	userWeight      uint64
}

func newVoteState(proposal ProposalInfo, voteKey string, voter std.Address, now uint64) (*VoteState, error) {
	cfg := GetConfigVersion(proposal.ConfigVersion)
	votingStartTime := proposal.State.CreatedAt + cfg.VotingStartDelay
	votingEndTime := votingStartTime + cfg.VotingPeriod

	hasUserVoted := false
	voted, exists := votes.Get(voteKey)
	if exists {
		hasUserVoted = voted.(bool)
	}

	return &VoteState{
		isVotingPeriod:  now >= votingStartTime && now < votingEndTime,
		isProposalValid: !proposal.State.Canceled,
		hasUserVoted:    hasUserVoted,
		userWeight:      proposal.PossibleAddressWithWeight[voter],
	}, nil
}

func (s *VoteState) validate() error {
	if !s.isVotingPeriod {
		return errUnableToVoteOutOfPeriod
	}
	if !s.isProposalValid {
		return errUnableToVoteCanceledProposal
	}
	if s.hasUserVoted {
		return errAlreadyVoted
	}
	if s.userWeight == 0 {
		return errNotEnoughVotingWeight
	}
	return nil
}

///////////////////// Vote Execution /////////////////////

type VoteExecutor struct {
	proposal *ProposalInfo
	voter    std.Address
	weight   uint64
}

func newVoteExecutor(proposal *ProposalInfo, voter std.Address, weight uint64) *VoteExecutor {
	return &VoteExecutor{
		proposal: proposal,
		voter:    voter,
		weight:   weight,
	}
}

func (e *VoteExecutor) execute(yes bool) error {
	if yes {
		e.proposal.Yea = safeVoteSum(e.proposal.Yea, e.weight)
	} else {
		e.proposal.Nay = safeVoteSum(e.proposal.Nay, e.weight)
	}
	return nil
}

func safeVoteSum(collected *u256.Uint, weight uint64) *u256.Uint {
	newSum, overflow := new(u256.Uint).AddOverflow(collected, u256.NewUint(weight))
	if overflow {
		panic(errOutOfRange)
	}
	return newSum
}

// Cancel cancels the proposal with the given ID.
// Only callable by the proposer or if the proposer's stake has fallen below the threshold others can call.
// ref: https://docs.gnoswap.io/contracts/governance/vote.gno#cancel
func Cancel(proposalId uint64) {
	common.IsHalted()

	en.MintAndDistributeGns()
	updateProposalsState()

	proposal := mustGetProposal(proposalId)
	if proposal.State.Canceled {
		panic(addDetailToError(
			errAlreadyCanceledProposal,
			ufmt.Sprintf("proposalId(%d) has already canceled", proposalId),
		))
	}

	config := GetConfigVersion(proposal.ConfigVersion)
	now := uint64(time.Now().Unix())
	if now >= (proposal.State.CreatedAt + config.VotingStartDelay) {
		panic(addDetailToError(
			errUnableToCancleVotingProposal,
			ufmt.Sprintf("voting has already started for proposalId(%d)", proposalId),
		))
	}

	caller := std.PrevRealm().Addr()
	if caller != proposal.Proposer {
		// If the caller is not the proposer, check if the proposer's stake has fallen below the threshold
		enough, balance, wanted := checkEnoughXGnsToPropose(proposal.Proposer)
		if enough {
			panic(addDetailToError(
				errUnableToCancelProposalWithVoterEnoughDelegated,
				ufmt.Sprintf(
					"caller(%s) is not the proposer(%s) and proposer's xgns balance(%d) is above the threshold(%d)",
					caller, proposal.Proposer,
					balance, wanted,
				),
			))
		}
	}

	proposal.State.Canceled = true
	proposal.State.CanceledAt = now
	proposal.State.Upcoming = false
	proposal.State.Active = false

	proposals.Set(formatUint(proposalId), proposal)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"Cancel",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"proposalId", formatUint(proposalId),
	)
}

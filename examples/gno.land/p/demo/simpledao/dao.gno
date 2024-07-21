package simpledao

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
	"gno.land/p/demo/ufmt"
)

var (
	ErrInvalidExecutor           = errors.New("invalid executor provided")
	ErrInsufficientProposalFunds = errors.New("insufficient funds for proposal")
	ErrInsufficientExecuteFunds  = errors.New("insufficient funds for executing proposal")
	ErrProposalExecuted          = errors.New("proposal already executed")
	ErrProposalInactive          = errors.New("proposal is inactive")
	ErrProposalNotAccepted       = errors.New("proposal is not accepted")
	ErrExecutorExpired           = errors.New("executor is expired")
)

var (
	minProposalFeeValue int64 = 100 * 1_000_000 // minimum gnot required for a govdao proposal (100 GNOT)
	minExecuteFeeValue  int64 = 500 * 1_000_000 // minimum gnot required for a govdao proposal (500 GNOT)

	minProposalFee = std.NewCoin("ugnot", minProposalFeeValue)
	minExecuteFee  = std.NewCoin("ugnot", minExecuteFeeValue)
)

// SimpleDAO is a simple DAO implementation
type SimpleDAO struct {
	proposals *avl.Tree // seqid.ID -> proposal
	membStore membstore.MemberStore
}

// New creates a new instance of the simpledao DAO
func New(membStore membstore.MemberStore) *SimpleDAO {
	return &SimpleDAO{
		proposals: avl.NewTree(),
		membStore: membStore,
	}
}

func (s *SimpleDAO) Propose(request dao.ProposalRequest) (uint64, error) {
	// Make sure the executor is set
	if request.Executor == nil {
		return 0, ErrInvalidExecutor
	}

	var (
		caller      = getDAOCaller()
		sentCoins   = std.GetOrigSend() // Get the sent coins, if any
		canCoverFee = sentCoins.AmountOf("ugnot") >= minProposalFee.Amount
	)

	// Check if the proposal is valid
	if !s.membStore.IsMember(caller) && !canCoverFee {
		return 0, ErrInsufficientProposalFunds
	}

	// Create the wrapped proposal
	prop := &proposal{
		author:                caller,
		description:           request.Description,
		executor:              request.Executor,
		status:                dao.Active,
		tally:                 newTally(),
		getTotalVotingPowerFn: s.membStore.TotalPower,
	}

	// Add the proposal
	id, err := s.addProposal(prop)
	if err != nil {
		return 0, ufmt.Errorf("unable to add proposal, %s", err.Error())
	}

	return id, nil
}

func (s *SimpleDAO) VoteOnProposal(id uint64, option dao.VoteOption) error {
	// Verify the GOVDAO member
	caller := getDAOCaller()

	member, err := s.membStore.Member(caller)
	if err != nil {
		return ufmt.Errorf("unable to get govdao member, %s", err.Error())
	}

	// Check if the proposal exists
	propRaw, err := s.ProposalByID(id)
	if err != nil {
		return ufmt.Errorf("unable to get proposal %d, %s", id, err.Error())
	}

	prop := propRaw.(*proposal)

	// Check the proposal status
	if prop.Status() == dao.ExecutionSuccessful ||
		prop.Status() == dao.ExecutionFailed {
		// Proposal was already executed, nothing to vote on anymore.
		//
		// In fact, the proposal should stop accepting
		// votes as soon as a 2/3+ majority is reached
		// on either option, but leaving the ability to vote still,
		// even if a proposal is accepted, or not accepted,
		// leaves room for "principle" vote decisions to be recorded
		return ErrProposalInactive
	}

	// Check if the executor is expired
	if prop.executor.IsExpired() {
		return ErrExecutorExpired
	}

	// Cast the vote
	if err = prop.tally.castVote(member, option); err != nil {
		return ufmt.Errorf("unable to vote on proposal %d, %s", id, err.Error())
	}

	// Emit the vote cast event
	std.Emit(
		dao.VoteAddedEvent,
		dao.VoteAddedIDKey, ufmt.Sprintf("%d", id),
		dao.VoteAddedAuthorKey, caller.String(),
		dao.VoteAddedOptionKey, option.String(),
	)

	// Check the votes to see if quorum is reached
	var (
		totalPower    = s.membStore.TotalPower()
		majorityPower = (2 * totalPower) / 3
	)

	acceptProposal := func() {
		prop.status = dao.Accepted

		std.Emit(
			dao.ProposalAcceptedEvent,
			dao.ProposalEventIDKey, ufmt.Sprintf("%d", id),
		)
	}

	declineProposal := func() {
		prop.status = dao.NotAccepted

		std.Emit(
			dao.ProposalNotAcceptedEvent,
			dao.ProposalEventIDKey, ufmt.Sprintf("%d", id),
		)
	}

	switch {
	case prop.tally.yays > majorityPower:
		// 2/3+ voted YES
		acceptProposal()
	case prop.tally.nays > majorityPower:
		// 2/3+ voted NO
		declineProposal()
	case prop.tally.abstains > majorityPower:
		// 2/3+ voted ABSTAIN
		declineProposal()
	case prop.tally.yays+prop.tally.nays+prop.tally.abstains >= totalPower:
		// Everyone voted, but it's undecided,
		// hence the proposal can't go through
		declineProposal()
	default:
		// Quorum not reached
	}

	return nil
}

func (s *SimpleDAO) ExecuteProposal(id uint64) error {
	var (
		caller      = getDAOCaller()
		sentCoins   = std.GetOrigSend() // Get the sent coins, if any
		canCoverFee = sentCoins.AmountOf("ugnot") >= minExecuteFee.Amount
	)

	// Check if the non-DAO member can cover the execute fee
	if !s.membStore.IsMember(caller) && !canCoverFee {
		return ErrInsufficientExecuteFunds
	}

	// Check if the proposal exists
	propRaw, err := s.ProposalByID(id)
	if err != nil {
		return ufmt.Errorf("unable to get proposal %d, %s", id, err.Error())
	}

	prop := propRaw.(*proposal)

	// Check if the proposal is executed
	if prop.Status() == dao.ExecutionSuccessful ||
		prop.Status() == dao.ExecutionFailed {
		// Proposal is already executed
		return ErrProposalExecuted
	}

	// Check the proposal status
	if prop.Status() != dao.Accepted {
		// Proposal is not accepted, cannot be executed
		return ErrProposalNotAccepted
	}

	// Check if the executor is expired
	if prop.executor.IsExpired() {
		return ErrExecutorExpired
	}

	// Emit an event when the execution finishes
	defer func() {
		std.Emit(
			dao.ProposalExecutedEvent,
			dao.ProposalEventIDKey, ufmt.Sprintf("%d", id),
			dao.ProposalEventExecutionKey, prop.status.String(),
		)
	}()

	// Attempt to execute the proposal
	if err = prop.executor.Execute(); err != nil {
		prop.status = dao.ExecutionFailed

		return ufmt.Errorf("error during proposal %d execution, %s", id, err.Error())
	}

	// Update the proposal status
	prop.status = dao.ExecutionSuccessful

	return nil
}

// getDAOCaller returns the DAO caller.
// XXX: This is not a great way to determine the caller, and it is very unsafe.
// However, the current MsgRun context does not persist escaping the main() scope.
// Until a better solution is developed, this enables proposals to be made through a package deployment + init()
func getDAOCaller() std.Address {
	return std.GetOrigCaller()
}

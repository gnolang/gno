package simpledao

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gov/dao"
)

var (
	errInvalidExecutor           = errors.New("invalid executor provided")
	errInsufficientProposalFunds = errors.New("insufficient funds for proposal")
	errInsufficientExecuteFunds  = errors.New("insufficient funds for executing proposal")
	errProposalExecuted          = errors.New("proposal already executed")
	errProposalInactive          = errors.New("proposal is inactive")
	errProposalNotAccepted       = errors.New("proposal is not accepted")
	errExecutorExpired           = errors.New("executor is expired")
	errNotGovDAO                 = errors.New("caller not correct govdao instance")
)

var (
	minProposalFeeValue int64 = 100 * 1_000_000 // minimum gnot required for a govdao proposal (100 GNOT)
	minExecuteFeeValue  int64 = 500 * 1_000_000 // minimum gnot required for a govdao proposal (500 GNOT)

	minProposalFee = std.NewCoin("ugnot", minProposalFeeValue)
	minExecuteFee  = std.NewCoin("ugnot", minExecuteFeeValue)
)

// SimpleDAO is a simple DAO implementation
type SimpleDAO struct {
	membStore *MembStore
	propStore *PropStore
}

// New creates a new instance of the simpledao DAO
func New(membStore *MembStore, propStore *PropStore) *SimpleDAO {
	return &SimpleDAO{
		membStore: membStore,
		propStore: propStore,
	}
}

func (s *SimpleDAO) Propose(description string, executor dao.Executor) (uint64, error) {
	// Make sure the executor is set
	if executor == nil {
		return 0, errInvalidExecutor
	}

	var (
		caller      = getDAOCaller()
		sentCoins   = std.GetOrigSend() // Get the sent coins, if any
		canCoverFee = sentCoins.AmountOf("ugnot") >= minProposalFee.Amount
	)

	// Check if the proposal is valid
	if !s.membStore.IsMember(caller) && !canCoverFee {
		return 0, errInsufficientProposalFunds
	}

	// Create the wrapped proposal
	prop := &proposal{
		author:      caller,
		description: description,
		executor:    executor,
		status:      dao.Active,
		votes:       newVotes(),
	}

	// Add the proposal
	id, err := s.propStore.addProposal(prop)
	if err != nil {
		return 0, ufmt.Errorf("unable to add proposal, %s", err.Error())
	}

	return id, nil
}

func (s *SimpleDAO) VoteOnProposal(id uint64, option dao.VoteOption) error {
	// Verify the GOVDAO member
	member, err := s.membStore.GetMember(getDAOCaller())
	if err != nil {
		return ufmt.Errorf("unable to get govdao member, %s", err.Error())
	}

	// Check if the proposal exists
	propRaw, err := s.propStore.GetProposalByID(id)
	if err != nil {
		return ufmt.Errorf("unable to get proposal %d, %s", id, err.Error())
	}

	prop := propRaw.(*proposal)

	// Check the proposal status
	if prop.GetStatus() == dao.ExecutionSuccessful ||
		prop.GetStatus() == dao.ExecutionFailed {
		// Proposal was already executed, nothing to vote on anymore.
		//
		// In fact, the proposal should stop accepting
		// votes as soon as a 2/3+ majority is reached
		// on either option, but leaving the ability to vote still,
		// even if a proposal is accepted, or not accepted,
		// leaves room for "principle" vote decisions to be recorded
		return errProposalInactive
	}

	// Check if the executor is expired
	if prop.executor.IsExpired() {
		return errExecutorExpired
	}

	// Cast the vote
	if err = prop.votes.castVote(member, option); err != nil {
		return ufmt.Errorf("unable to vote on proposal %d, %s", id, err.Error())
	}

	// Check the votes to see if quorum is reached
	var (
		majorityPower = s.membStore.getMajorityPower()
		totalPower    = s.membStore.getTotalPower()

		yays, nays, abstains = prop.votes.getTally()
	)

	switch {
	case yays > majorityPower:
		// 2/3+ voted YES
		prop.status = dao.Accepted
	case nays > majorityPower:
		// 2/3+ voted NO
		prop.status = dao.NotAccepted
	case abstains > majorityPower:
		// 2/3+ voted ABSTAIN
		prop.status = dao.NotAccepted
	case yays+nays+abstains >= totalPower:
		// Everyone voted, but it's undecided,
		// hence the proposal can't go through
		prop.status = dao.NotAccepted
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
		return errInsufficientExecuteFunds
	}

	// Check if the proposal exists
	propRaw, err := s.propStore.GetProposalByID(id)
	if err != nil {
		return ufmt.Errorf("unable to get proposal %d, %s", id, err.Error())
	}

	prop := propRaw.(*proposal)

	// Check if the proposal is executed
	if prop.GetStatus() == dao.ExecutionSuccessful ||
		prop.GetStatus() == dao.ExecutionFailed {
		// Proposal is already executed
		return errProposalExecuted
	}

	// Check the proposal status
	if prop.GetStatus() != dao.Accepted {
		// Proposal is not accepted, cannot be executed
		return errProposalNotAccepted
	}

	// Check if the executor is expired
	if prop.executor.IsExpired() {
		return errExecutorExpired
	}

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

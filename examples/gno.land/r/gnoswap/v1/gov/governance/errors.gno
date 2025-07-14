package governance

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

var (
	errOutOfRange                                     = errors.New("[GNOSWAP-GOVERNANCE-001] out of range for numeric value")
	errInvalidInput                                   = errors.New("[GNOSWAP-GOVERNANCE-002] invalid input")
	errDataNotFound                                   = errors.New("[GNOSWAP-GOVERNANCE-003] requested data not found")
	errNotEnoughBalance                               = errors.New("[GNOSWAP-GOVERNANCE-004] not enough balance")
	errUnableToVoteCanceledProposal                   = errors.New("[GNOSWAP-GOVERNANCE-005] unable to vote for canceled proposal")
	errAlreadyVoted                                   = errors.New("[GNOSWAP-GOVERNANCE-006] can not vote twice")
	errNotEnoughVotingWeight                          = errors.New("[GNOSWAP-GOVERNANCE-007] not enough voting power")
	errAlreadyCanceledProposal                        = errors.New("[GNOSWAP-GOVERNANCE-008] can not cancel already canceled proposal")
	errUnableToCancleVotingProposal                   = errors.New("[GNOSWAP-GOVERNANCE-009] unable to cancel voting proposal")
	errUnableToCancelProposalWithVoterEnoughDelegated = errors.New("[GNOSWAP-GOVERNANCE-010] unable to cancel proposal with voter has enough delegation")
	errTextProposalNotExecutable                      = errors.New("[GNOSWAP-GOVERNANCE-011] can not execute text proposal")
	errUnsupportedProposalType                        = errors.New("[GNOSWAP-GOVERNANCE-012] unsupported proposal type")
	errInvalidProposalType                            = errors.New("[GNOSWAP-GOVERNANCE-013] invalid proposal type")
	errUnableToVoteOutOfPeriod                        = errors.New("[GNOSWAP-GOVERNANCE-014] unable to vote out of voting period")
	errInvalidMessageFormat                           = errors.New("[GNOSWAP-GOVERNANCE-015] invalid message format")
	errProposalNotPassed                              = errors.New("[GNOSWAP-GOVERNANCE-016] proposal not passed")
	errInvalidAddress                                 = errors.New("[GNOSWAP-GOVERNANCE-017] invalid address")
	errExecutionWindowNotStarted                      = errors.New("[GNOSWAP-GOVERNANCE-018] execution window not started")
	errAlreadyActiveProposal                          = errors.New("[GNOSWAP-GOVERNANCE-019] already active proposal")
	errProposalNotFound                               = errors.New("[GNOSWAP-GOVERNANCE-020] proposal not found")
	errProposalNotExecutable                          = errors.New("[GNOSWAP-GOVERNANCE-021] proposal not executable")
)

func addDetailToError(err error, detail string) string {
	finalErr := ufmt.Errorf("%s || %s", err.Error(), detail)
	return finalErr.Error()
}

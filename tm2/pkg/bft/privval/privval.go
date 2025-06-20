package privval

import (
	"bytes"
	"fmt"

	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// PrivValidator signs votes and proposals for the consensus protocol using a
// signer (which can be either local or remote) and a state file to ensure that
// the validator never double sign, even in the case of a crash.
type PrivValidator struct {
	signer types.Signer
	state  *fstate.FileState
}

// PrivValidator errors.
var (
	errSameHRSBadData    = errors.New("same HRS with conflicting data")
	errSignatureMismatch = errors.New("state signature verification failed using signer public key")
)

// PrivValidator type implements types.PrivValidator.
var _ types.PrivValidator = (*PrivValidator)(nil)

// PubKey returns the public key of the private validator signer.
func (pv *PrivValidator) PubKey() crypto.PubKey {
	return pv.signer.PubKey()
}

// SignVote signs a vote using the private validator's signer and updates the
// state file to prevent double signing.
func (pv *PrivValidator) SignVote(chainID string, vote *types.Vote) error {
	// Check for identical height, round, step (HRS) against the last state.
	height, round, step := vote.Height, vote.Round, fstate.VoteTypeToStep(vote.Type)
	sameHRS, err := pv.state.CheckHRS(height, round, step)
	if err != nil {
		return err
	}

	// Get the bytes to sign from the vote.
	signBytes := vote.SignBytes(chainID)

	// We might crash before writing to the wal, causing us to try to re-sign
	// for the same HRS.
	if sameHRS {
		// If signBytes are the same, use the last signature.
		if bytes.Equal(signBytes, pv.state.SignBytes) {
			vote.Signature = pv.state.Signature
			return nil
		}

		// If they only differ by timestamp, use last timestamp and signature.
		if timestamp, ok := pv.state.CheckVotesOnlyDifferByTimestamp(signBytes); ok {
			vote.Signature = pv.state.Signature
			vote.Timestamp = timestamp
			return nil
		}

		// Otherwise, something is wrong.
		return errSameHRSBadData
	}

	// The HRS is different, so we need to sign the vote.
	signature, err := pv.signer.Sign(signBytes)
	if err != nil {
		return err
	}
	vote.Signature = signature

	// Then update the state and persist it.
	return pv.state.Update(height, round, step, signBytes, signature)
}

// SignProposal signs a proposal using the private validator's signer and updates the
// state file to prevent double signing.
func (pv *PrivValidator) SignProposal(chainID string, proposal *types.Proposal) error {
	// Check for identical height, round, step (HRS) against the last state.
	height, round, step := proposal.Height, proposal.Round, fstate.StepPropose
	sameHRS, err := pv.state.CheckHRS(height, round, step)
	if err != nil {
		return err
	}

	// Get the bytes to sign from the proposal.
	signBytes := proposal.SignBytes(chainID)

	// We might crash before writing to the wal, causing us to try to re-sign
	// for the same HRS.
	if sameHRS {
		// If signBytes are the same, use the last signature.
		if bytes.Equal(signBytes, pv.state.SignBytes) {
			proposal.Signature = pv.state.Signature
			return nil
		}

		// If they only differ by timestamp, use last timestamp and signature.
		if timestamp, ok := pv.state.CheckProposalsOnlyDifferByTimestamp(signBytes); ok {
			proposal.Signature = pv.state.Signature
			proposal.Timestamp = timestamp
			return nil
		}

		// Otherwise, something is wrong.
		return errSameHRSBadData
	}

	// The HRS is different, so we need to sign.
	signature, err := pv.signer.Sign(signBytes)
	if err != nil {
		return err
	}
	proposal.Signature = signature

	// Then update the state and persist it.
	return pv.state.Update(height, round, step, signBytes, signature)
}

// Close implements types.PrivValidator.
func (pv *PrivValidator) Close() error {
	return pv.signer.Close()
}

// PrivValidator type implements fmt.Stringer.
var _ fmt.Stringer = (*PrivValidator)(nil)

// String implements fmt.Stringer.
func (pv *PrivValidator) String() string {
	return fmt.Sprintf("PrivValidator{Signer: %v, State: %v}", pv.signer, pv.state)
}

// NewPrivValidator returns a new PrivValidator instance with the given signer and state
// file path. If the state file does not exist, it will be created.
func NewPrivValidator(signer types.Signer, stateFilePath string) (*PrivValidator, error) {
	// Load existing file state or create a new one.
	state, err := fstate.LoadOrMakeFileState(stateFilePath)
	if err != nil {
		return nil, err
	}

	// Check if validator state was signed by this signer.
	if state.SignBytes != nil {
		// Verify state signature using the signer public key.
		if !signer.PubKey().VerifyBytes(state.SignBytes, state.Signature) {
			return nil, errSignatureMismatch
		}
	}

	return &PrivValidator{
		signer: signer,
		state:  state,
	}, nil
}

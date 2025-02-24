package privval

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	rsclient "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/client"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
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
// An error is returned if the signer is remote and the connection fails.
func (pv *PrivValidator) PubKey() (crypto.PubKey, error) {
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
	// If the signer implements the io.Closer interface, close it.
	if closer, ok := pv.signer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
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
	state, err := fstate.NewFileState(stateFilePath)
	if err != nil {
		return nil, err
	}

	// Check if validator state was signed by this signer.
	if state.SignBytes != nil {
		// Get signer public key.
		pubKey, err := signer.PubKey()
		if err != nil {
			return nil, err
		}

		// Verify state signature using it.
		if !pubKey.VerifyBytes(state.SignBytes, state.Signature) {
			return nil, errSignatureMismatch
		}
	}

	return &PrivValidator{
		signer: signer,
		state:  state,
	}, nil
}

// NewPrivValidatorFromConfig returns a new PrivValidator instance based on the configuration.
// The clientLogger is only used for the remote signer client and ignored it the signer is local.
// The clientPrivKey is only used for the remote signer client using a TCP connection.
func NewPrivValidatorFromConfig(
	config *PrivValidatorConfig,
	clientPrivKey ed25519.PrivKeyEd25519,
	clientLogger *slog.Logger,
) (*PrivValidator, error) {
	var (
		signer types.Signer
		err    error
	)

	// Initialize the signer based on the configuration.
	// If the remote signer address is set, use a remote signer client.
	if config.RemoteSigner != nil && config.RemoteSigner.ServerAddress != "" {
		signer, err = rsclient.NewRemoteSignerClientFromConfig(
			config.RemoteSigner,
			clientPrivKey,
			clientLogger,
		)
	} else {
		// Otherwise, use a local signer.
		signer, err = local.NewLocalSigner(config.LocalSignerPath())
	}
	if err != nil {
		return nil, fmt.Errorf("signer initialization from config failed: %w", err)
	}

	return NewPrivValidator(signer, config.SignStatePath())
}

// Package common: HRSGuardedSigner — signer-side double-sign protection.
//
// A naive remote signer is dangerous: it relocates the *key* off the validator
// host, but leaves the *signing-state authority* on the validator. Any operator
// mistake that resets validator state (snapshot restore, disk reset,
// accidentally bringing two validators online) will then cause the signer to
// happily sign two conflicting messages at the same (height, round, step) and
// the validator gets slashed.
//
// HRSGuardedSigner closes that gap. It wraps an inner types.Signer with a
// FileState persisted on the *signer host* and refuses any Sign() request that
// is not strictly monotonic in (height, round, step) versus the prior call.
// Same-HRS replays of the exact same SignBytes idempotently return the cached
// signature; same-HRS with any non-byte-identical change is rejected.
//
// This mirrors tmkms's signer-side consensus.json gate.
package common

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// HRSGuardedSigner errors.
var (
	ErrUnparseableSignBytes = errors.New("hrs-guard: SignBytes are neither a CanonicalVote nor a CanonicalProposal")
	ErrSameHRSConflict      = errors.New("hrs-guard: same HRS with non-identical SignBytes")
)

// HRSGuardedSigner wraps an inner types.Signer with a persistent
// (height, round, step) gate. The state lives on the signer host; the gate
// fires before the inner signer is ever consulted.
type HRSGuardedSigner struct {
	inner  types.Signer
	state  *state.FileState
	logger *slog.Logger
	mu     sync.Mutex
}

// HRSGuardedSigner type implements types.Signer.
var _ types.Signer = (*HRSGuardedSigner)(nil)

// NewHRSGuardedSigner returns a guarded signer backed by a FileState at
// stateFilePath. The state file is created if it does not exist; if it
// exists, it is loaded and validated.
func NewHRSGuardedSigner(inner types.Signer, stateFilePath string, logger *slog.Logger) (*HRSGuardedSigner, error) {
	if inner == nil {
		return nil, errors.New("hrs-guard: inner signer is nil")
	}
	if stateFilePath == "" {
		return nil, errors.New("hrs-guard: state file path is empty")
	}

	fs, err := state.LoadOrMakeFileState(stateFilePath)
	if err != nil {
		return nil, fmt.Errorf("hrs-guard: load state: %w", err)
	}

	return &HRSGuardedSigner{
		inner:  inner,
		state:  fs,
		logger: logger,
	}, nil
}

// PubKey implements types.Signer.
func (g *HRSGuardedSigner) PubKey() crypto.PubKey {
	return g.inner.PubKey()
}

// Close implements types.Signer.
func (g *HRSGuardedSigner) Close() error {
	return g.inner.Close()
}

// classifySignBytes decodes signBytes as either a CanonicalVote or a
// CanonicalProposal and returns the consensus (height, round, step) triple.
// Returns ErrUnparseableSignBytes if the bytes match neither shape.
//
// Wire-type discrimination: CanonicalProposal has POLRound (fixed64) at field
// 4, while CanonicalVote has BlockID (length-delimited) at field 4. Decoding a
// proposal as a vote (or vice versa) typically fails at field 4 due to
// wire-type mismatch. The Type byte (PrevoteType=0x01, PrecommitType=0x02,
// ProposalType=0x20) gives a second consistency check.
func classifySignBytes(signBytes []byte) (height int64, round int, step state.Step, err error) {
	var vote types.CanonicalVote
	if vErr := amino.UnmarshalSized(signBytes, &vote); vErr == nil {
		if vote.Type == types.PrevoteType || vote.Type == types.PrecommitType {
			return vote.Height, int(vote.Round), state.VoteTypeToStep(vote.Type), nil
		}
	}

	var prop types.CanonicalProposal
	if pErr := amino.UnmarshalSized(signBytes, &prop); pErr == nil {
		if prop.Type == types.ProposalType {
			return prop.Height, int(prop.Round), state.StepPropose, nil
		}
	}

	return 0, 0, 0, ErrUnparseableSignBytes
}

// Sign implements types.Signer with HRS monotonicity enforcement.
//
// Sign refuses requests that regress (height, round, step) versus the
// persisted state. Same-HRS replay of identical SignBytes returns the cached
// signature idempotently; same-HRS with any byte difference is rejected as a
// double-sign attempt.
func (g *HRSGuardedSigner) Sign(signBytes []byte) ([]byte, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	height, round, step, err := classifySignBytes(signBytes)
	if err != nil {
		if g.logger != nil {
			g.logger.Warn("hrs-guard: refused unparseable SignBytes", "len", len(signBytes))
		}
		return nil, err
	}

	sameHRS, err := g.state.CheckHRS(height, round, step)
	if err != nil {
		if g.logger != nil {
			g.logger.Warn("hrs-guard: refused HRS regression",
				"request", fmt.Sprintf("H:%d R:%d S:%d", height, round, step),
				"last", g.state.String(),
				"error", err)
		}
		return nil, fmt.Errorf("hrs-guard: %w", err)
	}

	if sameHRS {
		// Same HRS as the last persisted sign. Allowed only if the bytes are
		// byte-identical (an upstream retransmit). Any deviation — including
		// timestamp-only — must be rejected here, because we cannot
		// communicate the canonical timestamp back to the validator and a
		// signature over old bytes will not verify against the validator's
		// new bytes.
		if bytes.Equal(signBytes, g.state.SignBytes) {
			return g.state.Signature, nil
		}
		if g.logger != nil {
			g.logger.Error("hrs-guard: refused same-HRS conflict",
				"hrs", fmt.Sprintf("H:%d R:%d S:%d", height, round, step))
		}
		return nil, ErrSameHRSConflict
	}

	// Strictly newer HRS. Sign, then persist before returning.
	signature, err := g.inner.Sign(signBytes)
	if err != nil {
		return nil, err
	}

	// If persistence fails we must NOT return the signature — that would
	// reproduce exactly the slashing failure mode this guard exists to
	// prevent (signed once, no record, signs again on next request).
	if err := g.state.Update(height, round, step, signBytes, signature); err != nil {
		if g.logger != nil {
			g.logger.Error("hrs-guard: state persist failed; refusing to release signature",
				"hrs", fmt.Sprintf("H:%d R:%d S:%d", height, round, step),
				"error", err)
		}
		return nil, fmt.Errorf("hrs-guard: persist state: %w", err)
	}

	return signature, nil
}

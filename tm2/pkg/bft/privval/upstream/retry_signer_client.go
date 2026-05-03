package upstream

// retry_signer_client.go: RetrySignerClient wraps SignerClient with
// retry semantics for transient errors.
//
// Direct port of cometbft/privval/retry_signer_client.go (CometBFT v0.39.1)
// adapted to tm2's PrivValidator interface (PubKey returns crypto.PubKey
// directly, no error — so the cached-pubkey pattern from SignerClient
// passes through unchanged).
//
// Retry policy: on transient errors (read/write timeout, no connection),
// sleep and retry up to N times. Non-transient errors — RemoteSignerError
// from tmkms (signer-side refusal: HRS regression, double-sign attempt,
// etc.) — pass through immediately. Retrying a refusal is wrong: it would
// turn an explicit safety abort into best-effort signing.

import (
	"errors"
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// RetrySignerClient wraps a SignerClient and retries on transient errors.
type RetrySignerClient struct {
	next    *SignerClient
	retries int
	timeout time.Duration
}

var _ types.PrivValidator = (*RetrySignerClient)(nil)

// NewRetrySignerClient returns a wrapper that retries each operation up
// to `retries` times with `timeout` sleep between attempts. If retries
// is 0, retries indefinitely (matches cometbft).
func NewRetrySignerClient(sc *SignerClient, retries int, timeout time.Duration) *RetrySignerClient {
	return &RetrySignerClient{next: sc, retries: retries, timeout: timeout}
}

func (rsc *RetrySignerClient) Close() error      { return rsc.next.Close() }
func (rsc *RetrySignerClient) IsConnected() bool { return rsc.next.IsConnected() }
func (rsc *RetrySignerClient) WaitForConnection(maxWait time.Duration) error {
	return rsc.next.WaitForConnection(maxWait)
}
func (rsc *RetrySignerClient) Init(maxWait time.Duration) error { return rsc.next.Init(maxWait) }
func (rsc *RetrySignerClient) Ping() error                      { return rsc.next.Ping() }

// PubKey returns the cached pubkey from the inner client. Init must
// have been called on the wrapped client (or via this wrapper's Init).
// No retry needed — the cached value has no failure mode.
func (rsc *RetrySignerClient) PubKey() crypto.PubKey {
	return rsc.next.PubKey()
}

// SignVote retries on transient errors. RemoteSignerError (signer-side
// refusal, e.g. HRS regression detected by tmkms's consensus.json gate)
// passes through immediately — retrying a slashing-prevention refusal
// would be a serious bug.
func (rsc *RetrySignerClient) SignVote(chainID string, vote *types.Vote) error {
	var err error
	for i := 0; i < rsc.retries || rsc.retries == 0; i++ {
		err = rsc.next.SignVote(chainID, vote)
		if err == nil {
			return nil
		}
		if !shouldRetry(err) {
			return err
		}
		time.Sleep(rsc.timeout)
	}
	return fmt.Errorf("upstream.RetrySignerClient: SignVote exhausted attempts: %w", err)
}

// SignProposal mirrors SignVote.
func (rsc *RetrySignerClient) SignProposal(chainID string, proposal *types.Proposal) error {
	var err error
	for i := 0; i < rsc.retries || rsc.retries == 0; i++ {
		err = rsc.next.SignProposal(chainID, proposal)
		if err == nil {
			return nil
		}
		if !shouldRetry(err) {
			return err
		}
		time.Sleep(rsc.timeout)
	}
	return fmt.Errorf("upstream.RetrySignerClient: SignProposal exhausted attempts: %w", err)
}

// shouldRetry decides whether an error is transient. Mirrors cometbft's
// "if RemoteSignerError, don't retry" check, plus tm2-specific transient
// errors (connection timeouts, no-conn).
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	var rse *WrappedRemoteSignerError
	if errors.As(err, &rse) {
		return false // signer explicitly refused; retrying would be wrong
	}
	if errors.Is(err, ErrConnectionTimeout) ||
		errors.Is(err, ErrReadTimeout) ||
		errors.Is(err, ErrWriteTimeout) ||
		errors.Is(err, ErrNoConnection) {
		return true
	}
	// Anything else (marshaling errors, decode errors, malformed responses)
	// is a programming/protocol bug, not a transient network issue. Don't
	// retry — surfacing the error fast helps debug.
	return false
}

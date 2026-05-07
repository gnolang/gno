package upstream

// signer_client.go: SignerClient implements types.PrivValidator over a
// SignerListenerEndpoint. Each PrivValidator method (PubKey, SignVote,
// SignProposal) is one request/response round-trip on the privval socket.
//
// Direct port of cometbft/privval/signer_client.go (CometBFT v0.39.1),
// with one structural adjustment for tm2's PrivValidator interface:
// PubKey() returns crypto.PubKey directly (no error), so we cache the
// validator's pubkey at construction time. The first PubKey() must
// follow a successful WaitForConnection — without it, PubKey() panics
// because the validator's identity isn't yet known.

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// SignerClient implements types.PrivValidator.
type SignerClient struct {
	endpoint *SignerListenerEndpoint
	chainID  string

	pubKeyMtx    sync.RWMutex
	cachedPubKey crypto.PubKey

	// verifiedGen is the endpoint connection generation at which the
	// signer's pubkey was last verified to match cachedPubKey. When the
	// endpoint reconnects (a new generation), the next sign call must
	// re-verify before signing — guards against a swap to a different
	// tmkms instance during a connection drop.
	verifiedGen atomic.Uint64
}

var _ types.PrivValidator = (*SignerClient)(nil)

// NewSignerClient constructs a client wrapping the given endpoint. The
// endpoint is started if it isn't running already. The validator's pubkey
// is fetched lazily via Init() (or implicitly on first SignVote) — see
// the doc on PubKey().
func NewSignerClient(endpoint *SignerListenerEndpoint, chainID string) (*SignerClient, error) {
	if !endpoint.IsRunning() {
		if err := endpoint.Start(); err != nil {
			return nil, fmt.Errorf("upstream.SignerClient: start endpoint: %w", err)
		}
	}
	return &SignerClient{endpoint: endpoint, chainID: chainID}, nil
}

// Init blocks for up to maxWait waiting for the signer to dial in, then
// fetches and caches the validator's consensus pubkey. Caller MUST call
// this (or otherwise populate the cache) before invoking PubKey() —
// the tm2 PrivValidator interface doesn't allow PubKey() to return an
// error, so we cache and panic on un-initialized access.
//
// Typical startup flow: construct client, call Init(maxWait); from then
// on PubKey()/SignVote/SignProposal are safe.
func (sc *SignerClient) Init(maxWait time.Duration) error {
	// Init is one-shot. cachedPubKey is the validator's committed
	// identity for this client; a second Init() against a swapped
	// signer would silently overwrite it, defeating the
	// verifyIdentityLocked invariant. The only legitimate caller is
	// node startup (privval/config.go), which calls Init exactly
	// once — a second call is a programmer bug, not a runtime
	// condition, so panic rather than return an error.
	sc.pubKeyMtx.RLock()
	already := sc.cachedPubKey != nil
	sc.pubKeyMtx.RUnlock()
	if already {
		panic("upstream.SignerClient: Init() called more than once — Init is one-shot")
	}

	if err := sc.endpoint.WaitForConnection(maxWait); err != nil {
		return fmt.Errorf("upstream.SignerClient: wait for signer: %w", err)
	}
	// Take the instance lock for the whole "fetch pubkey + record gen"
	// transaction so no reconnect can advance the gen between the fetch
	// and the record.
	sc.endpoint.Lock()
	defer sc.endpoint.Unlock()
	pk, err := sc.fetchPubKeyLocked()
	if err != nil {
		return err
	}
	sc.pubKeyMtx.Lock()
	sc.cachedPubKey = pk
	sc.pubKeyMtx.Unlock()
	sc.verifiedGen.Store(sc.endpoint.ConnectionGeneration())
	return nil
}

// Close shuts the endpoint down.
func (sc *SignerClient) Close() error {
	return sc.endpoint.Close()
}

// IsConnected reports whether the endpoint has a live conn to the signer.
func (sc *SignerClient) IsConnected() bool {
	return sc.endpoint.IsConnected()
}

// WaitForConnection blocks for up to maxWait waiting for the signer to
// connect. Equivalent to endpoint.WaitForConnection — exposed for
// callers that want to gate startup on signer availability.
func (sc *SignerClient) WaitForConnection(maxWait time.Duration) error {
	return sc.endpoint.WaitForConnection(maxWait)
}

// Ping sends a PingRequest. Used by callers that want to test the live
// connection without doing a sign request. CometBFT's pingLoop already
// handles application-level keepalive; this is a manual-poke alternative.
func (sc *SignerClient) Ping() error {
	resp, err := sc.endpoint.SendRequest(WrapMsg(&upstreampb.PingRequest{}))
	if err != nil {
		return err
	}
	if _, err := UnwrapMsg(resp); err != nil {
		return err
	}
	return nil
}

// PubKey returns the cached validator pubkey. Panics if not yet
// initialized — see Init().
func (sc *SignerClient) PubKey() crypto.PubKey {
	sc.pubKeyMtx.RLock()
	pk := sc.cachedPubKey
	sc.pubKeyMtx.RUnlock()
	if pk == nil {
		panic("upstream.SignerClient: PubKey() called before Init() — validator pubkey not yet known")
	}
	return pk
}

// SignVote sends the vote to the signer; on success only the Signature
// (and the canonicalized Timestamp) from the response are copied back
// into the caller's vote. Defense-in-depth: if a compromised or
// misbehaving signer returns a Vote with a different Height, Round,
// BlockID, or any other identifying field, we reject the response
// rather than letting the signer dictate WHAT we sign for. CometBFT's
// upstream `*vote = resp.Vote` convention conflates "trust the wire"
// with "trust the signer" — we don't.
//
// The whole sequence (identity re-verification + sign request) runs
// under the endpoint's instance lock so a reconnect can't substitute
// a different signer between the pubkey check and the vote signing.
func (sc *SignerClient) SignVote(chainID string, vote *types.Vote) error {
	if chainID != sc.chainID {
		return fmt.Errorf("upstream.SignerClient: chainID mismatch: got %q, client constructed for %q", chainID, sc.chainID)
	}
	pbVote, err := VoteToProto(vote)
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: VoteToProto: %w", err)
	}

	// Pre-flight: refuse fast if Init() was never called. Without
	// this, EnsureConnectionLocked below would block for the full
	// timeoutAccept on a never-Initialized client.
	if err := sc.requireInitialized(); err != nil {
		return err
	}

	sc.endpoint.Lock()
	defer sc.endpoint.Unlock()

	// Establish the conn FIRST. ensureConnection is what bumps the
	// connection-generation counter on a re-dial; calling it before
	// verifyIdentityLocked closes the TOCTOU window where a
	// DropConnection (which doesn't touch connGen) leaves the gen
	// short-circuit passing on stale state, then a fresh peer dials
	// in and would have its response accepted under the old cached
	// identity.
	if err := sc.endpoint.EnsureConnectionLocked(); err != nil {
		return fmt.Errorf("upstream.SignerClient: ensure connection: %w", err)
	}
	if err := sc.verifyIdentityLocked(); err != nil {
		return err
	}

	resp, err := sc.endpoint.SendRequestOnConnLocked(WrapMsg(&upstreampb.SignVoteRequest{
		Vote: pbVote, ChainId: chainID,
	}))
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: send: %w", err)
	}

	// Response-validation failures below indicate a malformed wire
	// message from the signer (corrupt envelope, wrong type,
	// proto-decodable but semantically broken Vote). The framing layer
	// has already accepted the bytes, so signer_endpoint won't drop the
	// conn — but reusing a conn that produced garbage is unsafe (the
	// next response could be the orphaned tail of the bad one). Drop
	// the conn here so the next sign call forces a fresh dial.
	// Exception: signed.Error is a legitimate refusal — the wire
	// envelope is well-formed and the signer is alive — so we leave
	// the conn intact for the next request.
	inner, err := UnwrapMsg(resp)
	if err != nil {
		sc.endpoint.DropConnection()
		return fmt.Errorf("upstream.SignerClient: unwrap: %w", err)
	}
	signed, ok := inner.(*upstreampb.SignedVoteResponse)
	if !ok {
		sc.endpoint.DropConnection()
		return fmt.Errorf("upstream.SignerClient: expected SignedVoteResponse, got %T", inner)
	}
	if signed.Error != nil {
		return &WrappedRemoteSignerError{Code: signed.Error.Code, Description: signed.Error.Description}
	}
	if signed.Vote == nil {
		sc.endpoint.DropConnection()
		return fmt.Errorf("upstream.SignerClient: SignedVoteResponse missing Vote")
	}

	signedVote, err := VoteFromProto(signed.Vote)
	if err != nil {
		sc.endpoint.DropConnection()
		return fmt.Errorf("upstream.SignerClient: VoteFromProto: %w", err)
	}
	if signedVote.Type != vote.Type ||
		signedVote.Height != vote.Height ||
		signedVote.Round != vote.Round ||
		signedVote.ValidatorIndex != vote.ValidatorIndex ||
		signedVote.ValidatorAddress != vote.ValidatorAddress ||
		!signedVote.BlockID.Equals(vote.BlockID) {
		sc.endpoint.DropConnection()
		return fmt.Errorf("upstream.SignerClient: signer echoed mismatched vote fields — refusing to use signature")
	}
	vote.Signature = signedVote.Signature
	vote.Timestamp = signedVote.Timestamp
	return nil
}

// SignProposal mirrors SignVote for proposals — same echo verification
// applies (signer may only fill in Signature and canonicalize Timestamp),
// and identity is re-verified atomically with the sign request.
func (sc *SignerClient) SignProposal(chainID string, proposal *types.Proposal) error {
	if chainID != sc.chainID {
		return fmt.Errorf("upstream.SignerClient: chainID mismatch: got %q, client constructed for %q", chainID, sc.chainID)
	}
	pbProp, err := ProposalToProto(proposal)
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: ProposalToProto: %w", err)
	}

	// Pre-flight: refuse fast if Init() was never called. Without
	// this, EnsureConnectionLocked below would block for the full
	// timeoutAccept on a never-Initialized client.
	if err := sc.requireInitialized(); err != nil {
		return err
	}

	sc.endpoint.Lock()
	defer sc.endpoint.Unlock()

	// See SignVote: ensureConnection BEFORE verifyIdentity to close
	// the DropConnection-doesn't-bump-gen TOCTOU window.
	if err := sc.endpoint.EnsureConnectionLocked(); err != nil {
		return fmt.Errorf("upstream.SignerClient: ensure connection: %w", err)
	}
	if err := sc.verifyIdentityLocked(); err != nil {
		return err
	}

	resp, err := sc.endpoint.SendRequestOnConnLocked(WrapMsg(&upstreampb.SignProposalRequest{
		Proposal: pbProp, ChainId: chainID,
	}))
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: send: %w", err)
	}

	// See SignVote: drop the conn on response-validation failures to
	// avoid reusing a conn that produced garbage. signed.Error is a
	// legitimate refusal — leave the conn intact.
	inner, err := UnwrapMsg(resp)
	if err != nil {
		sc.endpoint.DropConnection()
		return fmt.Errorf("upstream.SignerClient: unwrap: %w", err)
	}
	signed, ok := inner.(*upstreampb.SignedProposalResponse)
	if !ok {
		sc.endpoint.DropConnection()
		return fmt.Errorf("upstream.SignerClient: expected SignedProposalResponse, got %T", inner)
	}
	if signed.Error != nil {
		return &WrappedRemoteSignerError{Code: signed.Error.Code, Description: signed.Error.Description}
	}
	if signed.Proposal == nil {
		sc.endpoint.DropConnection()
		return fmt.Errorf("upstream.SignerClient: SignedProposalResponse missing Proposal")
	}

	signedProp, err := ProposalFromProto(signed.Proposal)
	if err != nil {
		sc.endpoint.DropConnection()
		return fmt.Errorf("upstream.SignerClient: ProposalFromProto: %w", err)
	}
	if signedProp.Type != proposal.Type ||
		signedProp.Height != proposal.Height ||
		signedProp.Round != proposal.Round ||
		signedProp.POLRound != proposal.POLRound ||
		!signedProp.BlockID.Equals(proposal.BlockID) {
		sc.endpoint.DropConnection()
		return fmt.Errorf("upstream.SignerClient: signer echoed mismatched proposal fields — refusing to use signature")
	}
	proposal.Signature = signedProp.Signature
	proposal.Timestamp = signedProp.Timestamp
	return nil
}

// requireInitialized returns a clear error if Init() has never been
// called, before any wire I/O is attempted. Without this short-
// circuit, SignVote/SignProposal would call EnsureConnectionLocked
// first and block for the full timeoutAccept on a never-Initialized
// client — caller would see "endpoint connection timed out" instead
// of the actionable "called before Init()" message.
func (sc *SignerClient) requireInitialized() error {
	sc.pubKeyMtx.RLock()
	cached := sc.cachedPubKey
	sc.pubKeyMtx.RUnlock()
	if cached == nil {
		return fmt.Errorf("upstream.SignerClient: SignVote/SignProposal called before Init() — refusing to sign without a cached identity")
	}
	return nil
}

// verifyIdentityLocked re-fetches the signer's pubkey and compares it
// against the cached identity if the endpoint has reconnected since the
// last verification. Caller MUST hold endpoint.Lock(). Returns an error
// if Init() was never called (no cached identity to verify against),
// the pubkey changed (refuse to sign for a swapped signer), or the
// re-fetch fails.
//
// The cachedPubKey nil-check runs FIRST, before the gen short-circuit.
// Otherwise, before any conn is established (currentGen == verifiedGen
// == 0), the gen check would return early and SignVote/SignProposal
// would proceed to sign without ever verifying the signer's identity
// against an expected pubkey — a defense-in-depth gap.
func (sc *SignerClient) verifyIdentityLocked() error {
	sc.pubKeyMtx.RLock()
	cached := sc.cachedPubKey
	sc.pubKeyMtx.RUnlock()
	if cached == nil {
		return fmt.Errorf("upstream.SignerClient: SignVote/SignProposal called before Init() — refusing to sign without a cached identity")
	}
	currentGen := sc.endpoint.ConnectionGeneration()
	if currentGen == sc.verifiedGen.Load() {
		return nil
	}
	pk, err := sc.fetchPubKeyLocked()
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: re-fetch pubkey on reconnect: %w", err)
	}
	if !pk.Equals(cached) {
		return fmt.Errorf("upstream.SignerClient: signer pubkey changed across reconnect (was %s, now %s) — refusing to sign",
			cached.Address(), pk.Address())
	}
	sc.verifiedGen.Store(currentGen)
	return nil
}

// fetchPubKeyLocked sends a PubKeyRequest and unwraps the response into
// a crypto.PubKey. Caller MUST hold endpoint.Lock() — used by both Init
// and verifyIdentityLocked, both of which need the lock for atomicity
// across the multi-RPC sequence.
func (sc *SignerClient) fetchPubKeyLocked() (crypto.PubKey, error) {
	resp, err := sc.endpoint.SendRequestLocked(WrapMsg(&upstreampb.PubKeyRequest{ChainId: sc.chainID}))
	if err != nil {
		return nil, fmt.Errorf("upstream.SignerClient: PubKeyRequest send: %w", err)
	}

	inner, err := UnwrapMsg(resp)
	if err != nil {
		return nil, err
	}
	pkResp, ok := inner.(*upstreampb.PubKeyResponse)
	if !ok {
		return nil, fmt.Errorf("upstream.SignerClient: expected PubKeyResponse, got %T", inner)
	}
	if pkResp.Error != nil {
		return nil, &WrappedRemoteSignerError{Code: pkResp.Error.Code, Description: pkResp.Error.Description}
	}
	return PubKeyFromProto(pkResp.PubKey)
}

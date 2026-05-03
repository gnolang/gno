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
	if err := sc.endpoint.WaitForConnection(maxWait); err != nil {
		return fmt.Errorf("upstream.SignerClient: wait for signer: %w", err)
	}
	pk, err := sc.fetchPubKey()
	if err != nil {
		return err
	}
	sc.pubKeyMtx.Lock()
	sc.cachedPubKey = pk
	sc.pubKeyMtx.Unlock()
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
	resp, err := sc.endpoint.SendRequest(*WrapMsg(&upstreampb.PingRequest{}))
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

// SignVote sends the vote to the signer; on success the response's
// signed Vote (Signature populated, possibly Timestamp canonicalized)
// replaces the input vote — matching cometbft/privval/signer_client.go's
// `*vote = resp.Vote` convention.
func (sc *SignerClient) SignVote(chainID string, vote *types.Vote) error {
	pbVote, err := VoteToProto(vote)
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: VoteToProto: %w", err)
	}

	resp, err := sc.endpoint.SendRequest(*WrapMsg(&upstreampb.SignVoteRequest{
		Vote: pbVote, ChainId: chainID,
	}))
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: send: %w", err)
	}

	inner, err := UnwrapMsg(resp)
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: unwrap: %w", err)
	}
	signed, ok := inner.(*upstreampb.SignedVoteResponse)
	if !ok {
		return fmt.Errorf("upstream.SignerClient: expected SignedVoteResponse, got %T", inner)
	}
	if signed.Error != nil {
		return &RemoteSignerErrorWrapper{Code: signed.Error.Code, Description: signed.Error.Description}
	}
	if signed.Vote == nil {
		return fmt.Errorf("upstream.SignerClient: SignedVoteResponse missing Vote")
	}

	signedVote, err := VoteFromProto(signed.Vote)
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: VoteFromProto: %w", err)
	}
	*vote = *signedVote
	return nil
}

// SignProposal mirrors SignVote for proposals.
func (sc *SignerClient) SignProposal(chainID string, proposal *types.Proposal) error {
	pbProp, err := ProposalToProto(proposal)
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: ProposalToProto: %w", err)
	}

	resp, err := sc.endpoint.SendRequest(*WrapMsg(&upstreampb.SignProposalRequest{
		Proposal: pbProp, ChainId: chainID,
	}))
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: send: %w", err)
	}

	inner, err := UnwrapMsg(resp)
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: unwrap: %w", err)
	}
	signed, ok := inner.(*upstreampb.SignedProposalResponse)
	if !ok {
		return fmt.Errorf("upstream.SignerClient: expected SignedProposalResponse, got %T", inner)
	}
	if signed.Error != nil {
		return &RemoteSignerErrorWrapper{Code: signed.Error.Code, Description: signed.Error.Description}
	}
	if signed.Proposal == nil {
		return fmt.Errorf("upstream.SignerClient: SignedProposalResponse missing Proposal")
	}

	signedProp, err := ProposalFromProto(signed.Proposal)
	if err != nil {
		return fmt.Errorf("upstream.SignerClient: ProposalFromProto: %w", err)
	}
	*proposal = *signedProp
	return nil
}

// fetchPubKey sends a PubKeyRequest and unwraps the response into a
// crypto.PubKey. Used by Init() to populate the cache; not exported
// because callers should go through Init().
func (sc *SignerClient) fetchPubKey() (crypto.PubKey, error) {
	resp, err := sc.endpoint.SendRequest(*WrapMsg(&upstreampb.PubKeyRequest{ChainId: sc.chainID}))
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
		return nil, &RemoteSignerErrorWrapper{Code: pkResp.Error.Code, Description: pkResp.Error.Description}
	}
	return PubKeyFromProto(pkResp.PubKey)
}

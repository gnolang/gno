package upstream_test

// signer_client_test.go: end-to-end exercise of SignerClient + RetrySignerClient
// against a fake signer. Verifies all PrivValidator interface methods (PubKey,
// SignVote, SignProposal) plus retry semantics on transient errors.

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePrivvalSigner: a tiny test double that handles PubKeyRequest,
// SignVoteRequest, SignProposalRequest with a real ed25519 key.
type fakePrivvalSigner struct {
	priv     ed25519.PrivKeyEd25519
	signFail bool // if true, return RemoteSignerError on Sign requests

	// tamperHeight, when non-zero, is written into the signed vote's
	// Height before responding — used to drive the echo-mismatch
	// rejection test.
	tamperHeight int64

	// Read-only after construction; copy if you need mutation.
	addr string
}

func newFakePrivvalSigner(t *testing.T, addr string) *fakePrivvalSigner {
	t.Helper()
	return &fakePrivvalSigner{
		priv: ed25519.GenPrivKey(),
		addr: addr,
	}
}

// serve loops accepting one inbound, handling N privval messages, and
// returning. Stops when ctx is done.
func (f *fakePrivvalSigner) serve(t *testing.T, ctx context.Context) {
	t.Helper()
	go func() {
		conn, err := net.Dial("tcp", f.addr)
		if err != nil {
			t.Errorf("fake: dial: %v", err)
			return
		}
		defer conn.Close()

		r := upstream.NewDelimitedReader(conn, upstream.MaxRemoteSignerMsgSize)
		w := upstream.NewDelimitedWriter(conn)

		for ctx.Err() == nil {
			var req upstreampb.Message
			if _, err := r.ReadMsg(&req); err != nil {
				return
			}

			inner, err := upstream.UnwrapMsg(&req)
			if err != nil {
				return
			}

			var resp interface{}
			switch m := inner.(type) {
			case *upstreampb.PingRequest:
				resp = &upstreampb.PingResponse{}
			case *upstreampb.PubKeyRequest:
				_ = m
				pbk, err := upstream.PubKeyToProto(f.priv.PubKey())
				if err != nil {
					t.Errorf("fake: PubKeyToProto: %v", err)
					return
				}
				resp = &upstreampb.PubKeyResponse{PubKey: pbk}
			case *upstreampb.SignVoteRequest:
				if f.signFail {
					resp = &upstreampb.SignedVoteResponse{
						Vote:  m.Vote,
						Error: &upstreampb.RemoteSignerError{Code: 1, Description: "test refusal"},
					}
					break
				}
				v := m.Vote
				if v != nil {
					v.Signature = []byte{0xde, 0xad, 0xbe, 0xef}
					if f.tamperHeight != 0 {
						v.Height = f.tamperHeight
					}
				}
				resp = &upstreampb.SignedVoteResponse{Vote: v}
			case *upstreampb.SignProposalRequest:
				p := m.Proposal
				if p != nil {
					p.Signature = []byte{0xca, 0xfe, 0xba, 0xbe}
				}
				resp = &upstreampb.SignedProposalResponse{Proposal: p}
			default:
				t.Errorf("fake: unhandled message type %T", inner)
				return
			}

			if _, err := w.WriteMsg(upstream.WrapMsg(resp)); err != nil {
				return
			}
		}
	}()
}

// startEndpoint creates a SignerListenerEndpoint on a fresh TCP port.
func startEndpoint(t *testing.T) (*upstream.SignerListenerEndpoint, net.Listener) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ep := upstream.NewSignerListenerEndpoint(logger, ln,
		upstream.SignerListenerEndpointTimeoutReadWrite(2*time.Second),
	)
	require.NoError(t, ep.Start())
	t.Cleanup(func() { _ = ep.Stop() })
	return ep, ln
}

// TestSignerClient_Init_FetchesPubKey: Init() blocks for the signer,
// then fetches and caches the validator's pubkey.
func TestSignerClient_Init_FetchesPubKey(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ep, ln := startEndpoint(t)
	signer := newFakePrivvalSigner(t, ln.Addr().String())
	signer.serve(t, ctx)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)

	require.NoError(t, sc.Init(3*time.Second))
	assert.Equal(t, signer.priv.PubKey().Bytes(), sc.PubKey().Bytes())
}

// TestSignerClient_PubKey_PanicsBeforeInit: tm2's PrivValidator interface
// can't return errors from PubKey(); SignerClient panics if used before
// Init().
func TestSignerClient_PubKey_PanicsBeforeInit(t *testing.T) {
	t.Parallel()
	ep, _ := startEndpoint(t)
	sc, err := upstream.NewSignerClient(ep, "test")
	require.NoError(t, err)

	assert.Panics(t, func() { _ = sc.PubKey() })
}

// TestSignerClient_SignVote_RoundTrip: Vote is sent, signer fills in a
// signature, response replaces the input vote.
func TestSignerClient_SignVote_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ep, ln := startEndpoint(t)
	signer := newFakePrivvalSigner(t, ln.Addr().String())
	signer.serve(t, ctx)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)
	require.NoError(t, sc.Init(3*time.Second))

	vote := &types.Vote{
		Type:             types.PrecommitType,
		Height:           42,
		Round:            3,
		ValidatorAddress: makeAddr(t, 0x55),
	}
	require.NoError(t, sc.SignVote("test-chain", vote))
	assert.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, vote.Signature)
}

// TestSignerClient_SignVote_RemoteError: signer-side refusal (e.g.
// HRS regression detected by tmkms's consensus.json gate) surfaces as
// *RemoteSignerErrorWrapper.
func TestSignerClient_SignVote_RemoteError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ep, ln := startEndpoint(t)
	signer := newFakePrivvalSigner(t, ln.Addr().String())
	signer.signFail = true
	signer.serve(t, ctx)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)
	require.NoError(t, sc.Init(3*time.Second))

	vote := &types.Vote{
		Type:             types.PrecommitType,
		Height:           42,
		Round:            3,
		ValidatorAddress: makeAddr(t, 0x00),
	}
	err = sc.SignVote("test-chain", vote)
	require.Error(t, err)

	rse := &upstream.RemoteSignerErrorWrapper{}
	require.ErrorAs(t, err, &rse)
	assert.EqualValues(t, 1, rse.Code)
	assert.Equal(t, "test refusal", rse.Description)
}

// TestSignerClient_SignProposal_RoundTrip: same shape as SignVote.
func TestSignerClient_SignProposal_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ep, ln := startEndpoint(t)
	signer := newFakePrivvalSigner(t, ln.Addr().String())
	signer.serve(t, ctx)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)
	require.NoError(t, sc.Init(3*time.Second))

	prop := &types.Proposal{
		Type:     types.ProposalType,
		Height:   100,
		Round:    2,
		POLRound: -1,
	}
	require.NoError(t, sc.SignProposal("test-chain", prop))
	assert.Equal(t, []byte{0xca, 0xfe, 0xba, 0xbe}, prop.Signature)
}

// TestSignerClient_SignVote_RejectsMismatchedEcho: a misbehaving (or
// compromised) signer that echoes a vote with mismatched fields must
// have its signature refused — we don't let the signer dictate WHAT
// gets signed, only the signature on what we asked it to sign.
func TestSignerClient_SignVote_RejectsMismatchedEcho(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ep, ln := startEndpoint(t)
	signer := newFakePrivvalSigner(t, ln.Addr().String())
	signer.tamperHeight = 9999 // signer rewrites Height in the response
	signer.serve(t, ctx)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)
	require.NoError(t, sc.Init(3*time.Second))

	vote := &types.Vote{
		Type:             types.PrecommitType,
		Height:           42,
		Round:            3,
		ValidatorAddress: makeAddr(t, 0x55),
	}
	err = sc.SignVote("test-chain", vote)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mismatched vote fields")
	// Caller's vote was NOT mutated to the tampered values.
	assert.EqualValues(t, 42, vote.Height)
	assert.Empty(t, vote.Signature, "tampered signature must not be copied back")
}

// TestSignerClient_PubKeyChangeOnReconnect_Rejected: if the held
// connection drops and a different signer dials in, the next SignVote
// must refuse rather than sign with a wrongly-attributed key. Threat
// model: an attacker who can force a TCP reset on the privval link and
// then race their own tmkms instance into the listener slot would
// otherwise be able to publish votes that the chain attributes to our
// validator (slashable behavior, identity drift).
func TestSignerClient_PubKeyChangeOnReconnect_Rejected(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ep, ln := startEndpoint(t)

	// Signer A dials in first.
	signerA := newFakePrivvalSigner(t, ln.Addr().String())
	signerA.serve(t, ctx)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)
	require.NoError(t, sc.Init(3*time.Second))
	pubA := signerA.priv.PubKey()
	require.Equal(t, pubA.Bytes(), sc.PubKey().Bytes())

	// Drop A's connection at the endpoint side, then have signer B (a
	// fresh keypair) dial in. This simulates a reconnect to a different
	// signer instance.
	ep.DropConnection()

	signerB := newFakePrivvalSigner(t, ln.Addr().String())
	signerB.serve(t, ctx)

	// Allow B's dial + handshake to complete. Don't sleep — poll briefly
	// until the endpoint advertises a new conn, then issue the SignVote.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if ep.IsConnected() {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	vote := &types.Vote{
		Type:             types.PrecommitType,
		Height:           42,
		Round:            3,
		ValidatorAddress: makeAddr(t, 0x77),
	}
	err = sc.SignVote("test-chain", vote)
	require.Error(t, err, "SignVote must refuse after pubkey-changing reconnect")
	assert.Contains(t, err.Error(), "pubkey changed across reconnect")
	assert.Empty(t, vote.Signature)
}

// TestRetrySignerClient_NoRetryOnRemoteError: signer-side refusal must
// pass through immediately. Retrying a slashing-prevention refusal would
// be a serious bug.
func TestRetrySignerClient_NoRetryOnRemoteError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ep, ln := startEndpoint(t)
	signer := newFakePrivvalSigner(t, ln.Addr().String())
	signer.signFail = true
	signer.serve(t, ctx)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)
	require.NoError(t, sc.Init(3*time.Second))

	rsc := upstream.NewRetrySignerClient(sc, 5, 10*time.Millisecond)

	vote := &types.Vote{
		Type:             types.PrecommitType,
		Height:           1,
		Round:            1,
		ValidatorAddress: makeAddr(t, 0xff),
	}
	start := time.Now()
	err = rsc.SignVote("test-chain", vote)
	elapsed := time.Since(start)

	require.Error(t, err)
	rse := &upstream.RemoteSignerErrorWrapper{}
	require.ErrorAs(t, err, &rse)
	// Should NOT have slept 5*10ms = 50ms — passed through immediately.
	assert.Less(t, elapsed, 30*time.Millisecond, "RemoteSignerError must pass through without retry sleep")
}

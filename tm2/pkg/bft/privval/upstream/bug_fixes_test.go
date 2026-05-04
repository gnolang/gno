package upstream_test

// bug_fixes_test.go: regression tests for bugs surfaced by an
// exploratory bug-hunt loop. Each test pins the post-fix behavior so
// the bug can't silently come back.

import (
	"bytes"
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

// ---- Bug #1 ------------------------------------------------------
//
// Stop() previously blocked for the full WaitForConnectionTimeout
// when called during a pending Init. WaitForConnection held
// instanceMtx through the wait, and OnStop tried to take the same
// lock; BaseService.Quit() is closed only AFTER OnStop returns, so
// adding <-Quit() to the wait wouldn't have helped. Fix: introduce
// a dedicated stopCh closed at the START of OnStop, release
// instanceMtx before the wait, and select on stopCh.

func TestBugFix_StopUnblocksPendingInit(t *testing.T) {
	t.Parallel()

	// Listen but never spawn a fake signer — Init's wait will time
	// out unless Stop unblocks it first.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ep := upstream.NewSignerListenerEndpoint(logger, ln,
		upstream.SignerListenerEndpointTimeoutReadWrite(2*time.Second),
	)
	require.NoError(t, ep.Start())

	sc, err := upstream.NewSignerClient(ep, "test")
	require.NoError(t, err)

	initDone := make(chan error, 1)
	go func() {
		initDone <- sc.Init(30 * time.Second) // generous budget
	}()

	// Give Init a moment to enter WaitForConnection.
	time.Sleep(50 * time.Millisecond)

	stopStart := time.Now()
	require.NoError(t, ep.Stop())
	stopDur := time.Since(stopStart)
	assert.Less(t, stopDur, 2*time.Second,
		"Stop must return promptly when Init is pending; took %v", stopDur)

	// Init must also return promptly with the timeout-shaped error
	// (we surface the stop as a connection-timeout to keep callers'
	// existing error handling working).
	select {
	case err := <-initDone:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wait for signer")
	case <-time.After(2 * time.Second):
		t.Fatal("Init did not return after Stop")
	}
}

// ---- Bug #2 ------------------------------------------------------
//
// SignVote / SignProposal previously completed without identity
// verification when Init() had not been called. verifyIdentityLocked
// short-circuited at currentGen == verifiedGen (both 0 before any
// conn) BEFORE checking cachedPubKey != nil. A caller that skipped
// Init would get a signature from a never-verified signer. Fix:
// move the cachedPubKey nil-check to the top of verifyIdentityLocked.

func TestBugFix_SignVoteWithoutInitRefuses(t *testing.T) {
	t.Parallel()

	// No fake signer — SignVote should refuse before any wire I/O,
	// so we don't need a peer to be dialing in.
	ep, _ := startEndpoint(t)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)

	vote := &types.Vote{
		Type:             types.PrecommitType,
		Height:           1,
		Round:            0,
		ValidatorAddress: makeAddr(t, 0xee),
	}
	err = sc.SignVote("test-chain", vote)
	require.Error(t, err, "SignVote without Init must refuse rather than sign")
	assert.Contains(t, err.Error(), "called before Init()")
	assert.Empty(t, vote.Signature, "vote must not be mutated when SignVote refuses")
}

func TestBugFix_SignProposalWithoutInitRefuses(t *testing.T) {
	t.Parallel()

	ep, _ := startEndpoint(t)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)

	prop := &types.Proposal{
		Type:     types.ProposalType,
		Height:   1,
		Round:    0,
		POLRound: -1,
	}
	err = sc.SignProposal("test-chain", prop)
	require.Error(t, err, "SignProposal without Init must refuse rather than sign")
	assert.Contains(t, err.Error(), "called before Init()")
	assert.Empty(t, prop.Signature)
}

// ---- Bug #4 ------------------------------------------------------
//
// SignVote / SendRequest after Stop previously blocked for the full
// timeoutAccept (default 3s). Bug #1's fix added stopCh to
// WaitForConnection, but ensureConnection (called from
// SendRequestLocked → sendRequestLocked) used base WaitConnection
// without a stopCh select arm. Fix: thread stopCh through
// signerEndpoint.WaitConnection so both WaitForConnection AND
// SendRequest unblock promptly when the endpoint is stopped.

func TestBugFix_SignVoteAfterStopReturnsPromptly(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ep, ln := startEndpoint(t)
	signer := newFakePrivvalSigner(t, ln.Addr().String())
	signer.serve(t, ctx)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)
	require.NoError(t, sc.Init(3*time.Second))

	require.NoError(t, ep.Stop())

	vote := &types.Vote{
		Type:             types.PrecommitType,
		Height:           1,
		Round:            0,
		ValidatorAddress: makeAddr(t, 0xff),
	}
	start := time.Now()
	err = sc.SignVote("test-chain", vote)
	dur := time.Since(start)

	require.Error(t, err, "SignVote on a stopped endpoint must error")
	assert.Less(t, dur, 500*time.Millisecond,
		"SignVote after Stop must return promptly via stopCh, not block for timeoutAccept (took %v)", dur)
}

// ---- Bug #3 ------------------------------------------------------
//
// signerEndpoint.ReadMessage / WriteMessage previously dropped the
// conn ONLY on timeoutError. On peer-initiated EOF (signer process
// exit, mid-handshake abort, network reset) the error was returned
// but se.conn stayed live. The next SendRequest saw IsConnected ==
// true, skipped ensureConnection's reconnect, and hit the dead conn
// again. Fix: drop the conn on ANY non-nil read/write error.

func TestBugFix_DropConnOnPeerEOF(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ep := upstream.NewSignerListenerEndpoint(logger, ln,
		upstream.SignerListenerEndpointTimeoutReadWrite(500*time.Millisecond),
	)
	require.NoError(t, ep.Start())
	t.Cleanup(func() { _ = ep.Stop() })

	// Peer dials and waits for a signal to close. This gives the
	// endpoint a live conn to install before we kill the peer.
	closePeer := make(chan struct{})
	go func() {
		conn, derr := net.Dial("tcp", ln.Addr().String())
		if derr != nil {
			return
		}
		<-closePeer
		_ = conn.Close()
	}()

	// Drive the endpoint to install the accepted conn. No SecretConn
	// wrapping is in use here, so the raw bytes go straight through.
	require.NoError(t, ep.WaitForConnection(2*time.Second))
	require.True(t, ep.IsConnected(), "endpoint should report connected after peer dial")

	// Peer closes the conn — simulates signer crash / network reset.
	close(closePeer)

	// Issue a request. Pre-fix: WriteMessage might succeed (TCP
	// buffers), ReadMessage would return EOF, conn would stay live,
	// and IsConnected would lie. Post-fix: any error drops the conn.
	_, err = ep.SendRequest(upstream.WrapMsg(&upstreampb.PingRequest{}))
	require.Error(t, err, "SendRequest into a peer-closed conn must error")

	assert.False(t, ep.IsConnected(),
		"after a peer-close + failed read/write, IsConnected must be false (conn dropped)")
}

// ---- Bug #6 ------------------------------------------------------
//
// verifyIdentityLocked had a TOCTOU window: DropConnection clears
// se.conn but does NOT bump connGen. That left the verify gen-check
// short-circuit (currentGen == verifiedGen) passing on stale state,
// after which SendRequestLocked → ensureConnection installed a fresh
// (potentially swapped) signer's conn — bumping connGen only AFTER
// the identity check had already passed. The vote then traveled down
// the unverified conn and the swapped signer's signature was copied
// into the caller's vote, attributed to the cached pubA identity.
//
// Fix: split SendRequestLocked into EnsureConnectionLocked +
// SendRequestOnConnLocked, and have SignerClient call them as
// ensure → verify → send. Identity verification now sees the
// up-to-date connGen produced by the fresh conn install.

// markedSigner: a fake privval signer with a caller-controlled
// signature marker, so the test can tell which signer produced a
// given signed vote.
type markedSigner struct {
	priv ed25519.PrivKeyEd25519
	sig  []byte
	addr string
}

func newMarkedSigner(t *testing.T, addr string, sig []byte) *markedSigner {
	t.Helper()
	return &markedSigner{
		priv: ed25519.GenPrivKey(),
		sig:  sig,
		addr: addr,
	}
}

func (m *markedSigner) serveOnce(t *testing.T, ctx context.Context, ready chan<- struct{}) {
	t.Helper()
	go func() {
		conn, err := net.Dial("tcp", m.addr)
		if err != nil {
			t.Logf("markedSigner: dial: %v", err)
			return
		}
		defer conn.Close()
		if ready != nil {
			close(ready)
		}

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
			switch req := inner.(type) {
			case *upstreampb.PingRequest:
				resp = &upstreampb.PingResponse{}
			case *upstreampb.PubKeyRequest:
				_ = req
				pbk, perr := upstream.PubKeyToProto(m.priv.PubKey())
				if perr != nil {
					return
				}
				resp = &upstreampb.PubKeyResponse{PubKey: pbk}
			case *upstreampb.SignVoteRequest:
				v := req.Vote
				if v != nil {
					v.Signature = append([]byte(nil), m.sig...)
				}
				resp = &upstreampb.SignedVoteResponse{Vote: v}
			default:
				return
			}
			if _, werr := w.WriteMsg(upstream.WrapMsg(resp)); werr != nil {
				return
			}
		}
	}()
}

// TestBugFix_VerifyIdentityRunsAfterEnsureConnection reproduces the
// TOCTOU race: signerA dials in and Init caches its key, then
// DropConnection clears the conn (but NOT connGen), then signerB
// dials in with a different key. Pre-fix: SignVote silently signed
// under signerB while sc.PubKey() still reported signerA. Post-fix:
// EnsureConnectionLocked runs first, bumping connGen to reflect
// signerB's conn, and verifyIdentityLocked then catches the swap and
// errors with "pubkey changed across reconnect".
func TestBugFix_VerifyIdentityRunsAfterEnsureConnection(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ep, ln := startEndpoint(t)

	sigA := []byte{0xAA, 0xAA, 0xAA, 0xAA}
	sigB := []byte{0xBB, 0xBB, 0xBB, 0xBB}

	// signerA dials in and caches its identity.
	signerA := newMarkedSigner(t, ln.Addr().String(), sigA)
	signerA.serveOnce(t, ctx, nil)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)
	require.NoError(t, sc.Init(3*time.Second))
	pubA := signerA.priv.PubKey()
	require.Equal(t, pubA.Bytes(), sc.PubKey().Bytes())
	gen1 := ep.ConnectionGeneration()

	// Simulate the post-pingLoop-timeout state: conn cleared, connGen
	// unchanged. (This is exactly what dropConnection() leaves behind
	// when ReadMessage hits EOF/timeout — see signer_endpoint.go.)
	ep.DropConnection()
	require.False(t, ep.IsConnected())
	require.Equal(t, gen1, ep.ConnectionGeneration(),
		"DropConnection must NOT bump connGen — that's what makes the TOCTOU window real")

	// signerB dials in with a different keypair.
	signerB := newMarkedSigner(t, ln.Addr().String(), sigB)
	pubB := signerB.priv.PubKey()
	require.NotEqual(t, pubA.Bytes(), pubB.Bytes())

	bReady := make(chan struct{})
	signerB.serveOnce(t, ctx, bReady)
	select {
	case <-bReady:
	case <-time.After(2 * time.Second):
		t.Fatal("signerB dial did not complete")
	}
	time.Sleep(100 * time.Millisecond) // let the listener enqueue B's conn

	// Drive a SignVote. With the fix, verifyIdentityLocked runs AFTER
	// ensureConnection installs signerB's conn, sees the new gen,
	// re-fetches pubkey, compares against cached pubA, and refuses.
	vote := &types.Vote{
		Type:             types.PrecommitType,
		Height:           42,
		Round:            3,
		ValidatorAddress: makeAddr(t, 0x99),
	}
	signErr := sc.SignVote("test-chain", vote)

	// Two acceptable post-fix outcomes:
	// (1) the conn that gets installed is signerB's, identity check
	//     sees the mismatch, SignVote errors.
	// (2) (rare) signerB's conn is dropped before install for some
	//     reason and SignVote errors with a connection error.
	// Either way, we MUST NOT see SignVote return success with sigB.
	require.Error(t, signErr, "SignVote must refuse after a swap; pre-fix it succeeded with signerB's signature")
	assert.NotEqual(t, sigB, vote.Signature,
		"vote.Signature must not contain signerB's marker — the swap must be rejected before the signature is copied back")

	// Cached identity must remain signerA's.
	assert.Equal(t, pubA.Bytes(), sc.PubKey().Bytes(),
		"PubKey() must continue to return signerA's key — cachedPubKey must not drift on a reconnect race")

	// Mute unused-helper warnings if the test infra changes.
	_ = bytes.Equal
}

// ---- Bug #7 ------------------------------------------------------
//
// Init() previously overwrote cachedPubKey unconditionally on every
// call. A second Init() against a swapped signer (different keypair)
// silently replaced the validator's committed identity, defeating the
// verifyIdentityLocked invariant that anchors all subsequent
// SignVote / SignProposal calls — pubkey-swap detection is
// meaningless if Init can quietly re-anchor to the new key. Fix:
// reject the second Init() outright.

func TestBugFix_InitRefusesSecondCall(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ep, ln := startEndpoint(t)

	sigA := []byte{0xAA, 0xAA, 0xAA, 0xAA}
	signerA := newMarkedSigner(t, ln.Addr().String(), sigA)
	signerA.serveOnce(t, ctx, nil)

	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)
	require.NoError(t, sc.Init(3*time.Second))
	pubA := signerA.priv.PubKey()
	require.Equal(t, pubA.Bytes(), sc.PubKey().Bytes())

	// Drop the conn and let signerB take its place — exactly the
	// hostile scenario where unguarded Init would re-anchor.
	ep.DropConnection()
	signerB := newMarkedSigner(t, ln.Addr().String(), []byte{0xBB, 0xBB, 0xBB, 0xBB})
	bReady := make(chan struct{})
	signerB.serveOnce(t, ctx, bReady)
	select {
	case <-bReady:
	case <-time.After(2 * time.Second):
		t.Fatal("signerB dial did not complete")
	}

	// Second Init() must refuse outright — never replace cachedPubKey.
	err = sc.Init(3 * time.Second)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already called")
	assert.Equal(t, pubA.Bytes(), sc.PubKey().Bytes(),
		"cachedPubKey must remain signerA's key — second Init must not replace it")
}

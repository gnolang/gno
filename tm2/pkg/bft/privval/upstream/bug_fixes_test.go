package upstream_test

// bug_fixes_test.go: regression tests for three bugs surfaced by an
// exploratory bug-hunt loop. Each test pins the post-fix behavior so
// the bug can't silently come back.

import (
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
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

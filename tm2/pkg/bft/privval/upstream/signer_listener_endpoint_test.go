package upstream_test

// signer_listener_endpoint_test.go: end-to-end smoke test for the
// SignerListenerEndpoint. A fake signer dials in, we exchange a
// PubKeyRequest/Response, and verify the listener wires it through.
//
// Modeled on cometbft/privval/signer_listener_endpoint_test.go.

import (
	"context"
	"crypto/rand"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSigner is a tiny test double for tmkms: dials in, reads one
// message, responds with PubKeyResponse carrying a fixed ed25519 pubkey.
// Used to verify the listener exchanges messages correctly.
type fakeSigner struct {
	addr      string
	pubKeyEd  []byte // 32 bytes
	respondCh chan struct{}
	doneCh    chan struct{}
}

func newFakeSigner(t *testing.T, addr string) *fakeSigner {
	t.Helper()
	pk := make([]byte, 32)
	_, err := rand.Read(pk)
	require.NoError(t, err)
	return &fakeSigner{
		addr:      addr,
		pubKeyEd:  pk,
		respondCh: make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

// connect dials the listener address, reads one privval message, sends
// a PubKeyResponse with the fake's pubkey, then closes the conn. Runs in
// a goroutine.
func (f *fakeSigner) connect(t *testing.T, ctx context.Context) {
	t.Helper()
	go func() {
		defer close(f.doneCh)

		var conn net.Conn
		// Retry briefly while listener is coming up.
		deadline := time.Now().Add(2 * time.Second)
		for {
			c, err := net.Dial("tcp", f.addr)
			if err == nil {
				conn = c
				break
			}
			if time.Now().After(deadline) {
				t.Errorf("fakeSigner: dial failed: %v", err)
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		defer conn.Close()

		// Read one request from the listener.
		r := upstream.NewDelimitedReader(conn, upstream.MaxRemoteSignerMsgSize)
		var msg upstreampb.Message
		if _, err := r.ReadMsg(&msg); err != nil && err != io.EOF {
			t.Errorf("fakeSigner: ReadMsg: %v", err)
			return
		}

		// Wait for the test to grant permission to respond (so we can
		// observe the listener's "blocked on response" state).
		select {
		case <-f.respondCh:
		case <-ctx.Done():
			return
		}

		// Reply with a PubKeyResponse containing our fake ed25519 pubkey.
		resp := upstream.WrapMsg(&upstreampb.PubKeyResponse{
			PubKey: &upstreampb.PublicKey{
				Sum: &upstreampb.PublicKey_Ed25519{Ed25519: f.pubKeyEd},
			},
		})
		w := upstream.NewDelimitedWriter(conn)
		if _, err := w.WriteMsg(resp); err != nil {
			t.Errorf("fakeSigner: WriteMsg: %v", err)
		}
	}()
}

// TestSignerListenerEndpoint_SendRequest_RoundTrip exercises the listener
// end-to-end: bind, accept inbound, exchange a PubKeyRequest, get the
// expected PubKeyResponse out.
func TestSignerListenerEndpoint_SendRequest_RoundTrip(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Bind the listener.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	// Construct the endpoint over the raw listener (no SecretConnection
	// for this smoke test — wire compat is verified by msgs_test.go).
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ep := upstream.NewSignerListenerEndpoint(logger, ln,
		upstream.SignerListenerEndpointTimeoutReadWrite(2*time.Second),
	)
	require.NoError(t, ep.Start())
	t.Cleanup(func() { _ = ep.Stop() })

	// Spin up a fake signer that will dial in.
	signer := newFakeSigner(t, ln.Addr().String())
	signer.connect(t, ctx)

	// Wait for the signer to connect.
	require.NoError(t, ep.WaitForConnection(3*time.Second))

	// Send a PubKeyRequest, allow the signer to respond, expect a matching
	// PubKeyResponse back.
	go func() {
		// Tell the fakeSigner it can respond now (request is in flight).
		// Small sleep to ensure the listener has dispatched the write
		// before we authorize the response. Belt-and-suspenders for
		// race-detector cleanliness; SendRequest blocks on read so the
		// ordering is naturally correct.
		time.Sleep(50 * time.Millisecond)
		close(signer.respondCh)
	}()

	resp, err := ep.SendRequest(*upstream.WrapMsg(&upstreampb.PubKeyRequest{ChainId: "test"}))
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Unwrap and check.
	inner, err := upstream.UnwrapMsg(resp)
	require.NoError(t, err)
	pkr, ok := inner.(*upstreampb.PubKeyResponse)
	require.True(t, ok, "expected PubKeyResponse, got %T", inner)
	assert.Equal(t, signer.pubKeyEd, pkr.PubKey.GetEd25519())

	// Wait for the fake to finish.
	select {
	case <-signer.doneCh:
	case <-ctx.Done():
		t.Fatal("fakeSigner did not finish")
	}
}

// TestSignerListenerEndpoint_WaitForConnection_Timeout: with no signer
// dialing in, WaitForConnection returns ErrConnectionTimeout.
func TestSignerListenerEndpoint_WaitForConnection_Timeout(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ep := upstream.NewSignerListenerEndpoint(logger, ln)
	require.NoError(t, ep.Start())
	t.Cleanup(func() { _ = ep.Stop() })

	err = ep.WaitForConnection(200 * time.Millisecond)
	require.ErrorIs(t, err, upstream.ErrConnectionTimeout)
}

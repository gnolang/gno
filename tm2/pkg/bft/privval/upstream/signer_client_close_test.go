package upstream_test

// signer_client_close_test.go pins PR #5717 review finding 3 (the cross-cutting
// MAJOR also flagged on #5718): SignerClient.Close must stop the endpoint's
// service — closing the listener, ping loop and service loop and releasing the
// bound port — not merely DropConnection. Node shutdown closes the validator
// via PrivValidator.Close, so a Close that only drops the conn leaks the
// listener and goroutines and holds the port, blocking clean in-process
// restart.

import (
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream"
	"github.com/stretchr/testify/require"
)

func TestSignerClient_Close_StopsEndpointAndReleasesPort(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ep := upstream.NewSignerListenerEndpoint(logger, ln,
		upstream.SignerListenerEndpointTimeoutReadWrite(2*time.Second),
	)

	// NewSignerClient starts the endpoint (serviceLoop + pingLoop, listening).
	sc, err := upstream.NewSignerClient(ep, "test-chain")
	require.NoError(t, err)
	require.True(t, ep.IsRunning(), "endpoint should be running after NewSignerClient")

	require.NoError(t, sc.Close())

	// Close must stop the endpoint's service, not just drop the conn.
	require.False(t, ep.IsRunning(), "Close must stop the endpoint service")

	// ...and release the bound port. On the pre-fix code the listener stayed
	// open and this re-bind fails with EADDRINUSE.
	ln2, err := net.Listen("tcp", addr)
	require.NoError(t, err, "Close must release the listener port")
	_ = ln2.Close()

	// Close is idempotent: a second call is a no-op, not an error.
	require.NoError(t, sc.Close())
}

package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	r "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote"
	c "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/client"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	unixSocketPath = "/tmp/test_tm2_remote_signer"
	tcpLocalhost   = "tcp://127.0.0.1"
)

func testUnixSocket(t *testing.T) string {
	t.Helper()

	// Ensure the unix socket path exists.
	if err := os.MkdirAll(unixSocketPath, 0o755); err != nil {
		t.Fatalf("failed to create unix socket path: %v", err)
	}

	// Create a unique unix socket file path.
	filePath := fmt.Sprintf("%s/%s.sock", unixSocketPath, xid.New().String())

	// Ensure the file is deleted after the test.
	t.Cleanup(func() {
		os.Remove(filePath)
	})

	return fmt.Sprintf("unix://%s", filePath)
}

func newRemoteSignerClient(t *testing.T, address string) *c.RemoteSignerClient {
	t.Helper()

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	rsc, _ := c.NewRemoteSignerClient(
		ctx,
		address,
		log.NewNoopLogger(),
	)

	return rsc
}

func newRemoteSignerServer(t *testing.T, address string, signer types.Signer) *RemoteSignerServer {
	t.Helper()

	if signer == nil {
		signer = types.NewMockSigner()
	}

	rss, _ := NewRemoteSignerServer(
		signer,
		address,
		log.NewNoopLogger(),
	)

	return rss
}

func TestCloseState(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		// Init a new remote signer client without connection.
		rss := newRemoteSignerServer(t, testUnixSocket(t), nil)
		require.False(t, rss.IsRunning())

		// Try to stop it.
		require.ErrorIs(t, rss.Stop(), ErrServerAlreadyStopped)
		require.False(t, rss.IsRunning())

		// Start it.
		require.NoError(t, rss.Start())
		require.True(t, rss.IsRunning())

		// Start it again.
		require.ErrorIs(t, rss.Start(), ErrServerAlreadyStarted)
		require.True(t, rss.IsRunning())

		// Stop it.
		require.NoError(t, rss.Stop())
		assert.False(t, rss.IsRunning())
	})

	t.Run("listeners cleanup", func(t *testing.T) {
		t.Parallel()

		unixSocket := testUnixSocket(t)

		// Init a new remote signer server.
		rss, err := NewRemoteSignerServer(types.NewMockSigner(), unixSocket, log.NewNoopLogger())
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())

		// Stop the server then Start it again.
		require.NoError(t, rss.Stop())
		require.NoError(t, rss.Start())
		require.NotNil(t, rss.listener)

		// Stop the server with the listener already closed.
		rss.listener.Close()
		require.Error(t, rss.Stop())

		// Start it again with a faulty listener.
		_, address := osm.ProtocolAndAddress(unixSocket)
		file, err := os.Create(address)
		require.NotNil(t, file)
		require.NoError(t, err)
		err = file.Chmod(0o000)
		require.NoError(t, err)
		assert.ErrorIs(t, rss.Start(), ErrListenFailed)
	})
}

func TestServerResponse(t *testing.T) {
	t.Parallel()

	t.Run("PubKey response", func(t *testing.T) {
		t.Parallel()

		var (
			unixSocket = testUnixSocket(t)
			signer     = types.NewMockSigner()
		)

		// Test a valid PubKey response.
		rss := newRemoteSignerServer(t, unixSocket, signer)
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		rsc := newRemoteSignerClient(t, unixSocket)
		require.NotNil(t, rsc)
		remotePK := rsc.PubKey()
		require.NotNil(t, remotePK)
		localPK := signer.PubKey()
		require.NotNil(t, localPK)
		require.Equal(t, localPK, remotePK)
		rss.Stop()
		rsc.Close()
	})

	t.Run("Sign response", func(t *testing.T) {
		t.Parallel()

		var (
			unixSocket = testUnixSocket(t)
			signer     = types.NewMockSigner()
		)

		// Test a valid Sign response.
		rss := newRemoteSignerServer(t, unixSocket, signer)
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		rsc := newRemoteSignerClient(t, unixSocket)
		require.NotNil(t, rsc)
		remoteSignature, err := rsc.Sign([]byte("sign bytes"))
		require.NotNil(t, remoteSignature)
		require.NoError(t, err)
		localSignature, err := signer.Sign([]byte("sign bytes"))
		require.NotNil(t, localSignature)
		require.NoError(t, err)
		require.Equal(t, localSignature, remoteSignature)
		rss.Stop()
		rsc.Close()

		// Test an erroring Sign response.
		signer = types.NewErroringMockSigner()
		rss = newRemoteSignerServer(t, unixSocket, signer)
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		rsc = newRemoteSignerClient(t, unixSocket)
		require.NotNil(t, rsc)
		remoteSignature, err = rsc.Sign([]byte("sign bytes"))
		require.Nil(t, remoteSignature)
		require.Error(t, err)
		localSignature, err = signer.Sign([]byte("sign bytes"))
		require.Nil(t, localSignature)
		assert.Error(t, err)
		rss.Stop()
		rsc.Close()
	})

	t.Run("Invalid request", func(t *testing.T) {
		t.Parallel()

		rss := newRemoteSignerServer(t, testUnixSocket(t), types.NewMockSigner())
		assert.Nil(t, rss.handleRequest([]byte("invalid request")))
	})
}

func TestServerConnection(t *testing.T) {
	t.Parallel()

	t.Run("tcp configuration succeeded", func(t *testing.T) {
		t.Parallel()

		// Client succeeded authenticating server.
		clientPrivKey := ed25519.GenPrivKey()
		rss, err := NewRemoteSignerServer(
			types.NewMockSigner(),
			tcpLocalhost+":0",
			log.NewNoopLogger(),
			WithAuthorizedKeys([]ed25519.PubKeyEd25519{clientPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		serverPort := rss.ListenAddress(t).(*net.TCPAddr).Port

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()
		rsc, err := c.NewRemoteSignerClient(
			ctx,
			fmt.Sprintf("%s:%d", tcpLocalhost, serverPort),
			log.NewNoopLogger(),
			c.WithClientPrivKey(clientPrivKey),
		)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		_, err = rsc.Sign([]byte("test"))
		require.NoError(t, err)
		rss.Stop()
		rsc.Close()

		// Server succeeded authenticating client.
		serverPrivKey := ed25519.GenPrivKey()
		rss, err = NewRemoteSignerServer(
			types.NewMockSigner(),
			tcpLocalhost+":0",
			log.NewNoopLogger(),
			WithServerPrivKey(serverPrivKey),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		serverPort = rss.ListenAddress(t).(*net.TCPAddr).Port

		ctx, cancelFn = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()
		rsc, err = c.NewRemoteSignerClient(
			ctx,
			fmt.Sprintf("%s:%d", tcpLocalhost, serverPort),
			log.NewNoopLogger(),
			c.WithAuthorizedKeys([]ed25519.PubKeyEd25519{serverPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		_, err = rsc.Sign([]byte("test"))
		assert.NoError(t, err)
		rss.Stop()
		rsc.Close()
	})

	t.Run("tcp configuration failed", func(t *testing.T) {
		t.Parallel()

		// Client fails authenticating server.
		rss, err := NewRemoteSignerServer(
			types.NewMockSigner(),
			tcpLocalhost+":0",
			log.NewNoopLogger(),
			WithAuthorizedKeys([]ed25519.PubKeyEd25519{ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		serverPort := rss.ListenAddress(t).(*net.TCPAddr).Port

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()
		rsc, err := c.NewRemoteSignerClient(
			ctx,
			fmt.Sprintf("%s:%d", tcpLocalhost, serverPort),
			log.NewNoopLogger(),
			c.WithAuthorizedKeys([]ed25519.PubKeyEd25519{ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.Nil(t, rsc)
		assert.ErrorIs(t, err, r.ErrUnauthorizedPubKey)
		rss.Stop()
	})
}

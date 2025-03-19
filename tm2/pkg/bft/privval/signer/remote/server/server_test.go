package server

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
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

	return fmt.Sprintf("unix://%s/%s.sock", unixSocketPath, xid.New().String())
}

func newRemoteSignerClient(t *testing.T, address string) *c.RemoteSignerClient {
	t.Helper()

	rsc, _ := c.NewRemoteSignerClient(
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
		[]string{address},
		log.NewNoopLogger(),
	)

	return rss
}

func TestCloseState(t *testing.T) {
	t.Parallel()

	// Create a directory for the unix socket.
	os.MkdirAll(unixSocketPath, 0o755)

	// Remove the directory after the test.
	t.Cleanup(func() {
		os.Remove(unixSocketPath)
	})

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

		// Wait should not block.
		rss.Wait()
	})

	t.Run("listeners cleanup", func(t *testing.T) {
		t.Parallel()

		unixSockets := []string{
			testUnixSocket(t),
			testUnixSocket(t),
			testUnixSocket(t),
		}

		// Init a new remote signer server.
		rss, err := NewRemoteSignerServer(types.NewMockSigner(), unixSockets, log.NewNoopLogger())
		require.Len(t, rss.listeners, 3)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())

		// Stop the server then Start it again.
		require.NoError(t, rss.Stop())
		for _, listener := range rss.listeners {
			require.Nil(t, listener)
		}
		require.NoError(t, rss.Start())
		for _, listener := range rss.listeners {
			require.NotNil(t, listener)
		}

		// Stop the server with a listener already closed.
		rss.listeners[0].Close()
		require.Error(t, rss.Stop())

		// Start it again with a faulty listener.
		_, address := osm.ProtocolAndAddress(unixSockets[1])
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

	// Create a directory for the unix socket.
	os.MkdirAll(unixSocketPath, 0o755)

	// Remove the directory after the test.
	t.Cleanup(func() {
		os.Remove(unixSocketPath)
	})

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
		remotePK, err := rsc.PubKey()
		require.NotNil(t, remotePK)
		require.NoError(t, err)
		localPK, err := signer.PubKey()
		require.NotNil(t, localPK)
		require.NoError(t, err)
		require.Equal(t, localPK, remotePK)
		rss.Stop()
		rsc.Close()

		// Test an erroring PubKey response.
		signer = types.NewErroringMockSigner()
		rss = newRemoteSignerServer(t, unixSocket, signer)
		require.NotNil(t, rss)
		require.NoError(t, rss.Start())
		rsc = newRemoteSignerClient(t, unixSocket)
		require.NotNil(t, rsc)
		remotePK, err = rsc.PubKey()
		require.Nil(t, remotePK)
		require.Error(t, err)
		localPK, err = signer.PubKey()
		require.Nil(t, localPK)
		assert.Error(t, err)
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

	t.Run("Ping response", func(t *testing.T) {
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
		assert.NoError(t, rsc.Ping())
		rss.Stop()
		rsc.Close()
	})

	t.Run("Invalid request", func(t *testing.T) {
		t.Parallel()

		rss := newRemoteSignerServer(t, testUnixSocket(t), types.NewMockSigner())
		assert.Nil(t, rss.handle([]byte("invalid request")))
	})
}

func TestServerConnection(t *testing.T) {
	t.Parallel()

	// Create a directory for the unix socket.
	os.MkdirAll(unixSocketPath, 0o755)

	// Remove the directory after the test.
	t.Cleanup(func() {
		os.Remove(unixSocketPath)
	})

	t.Run("conn closed during read/write", func(t *testing.T) {
		t.Parallel()

		// Server that fails on read.
		newReadWriteErrorRemoteSignerClient := func(t *testing.T, address string, noWrite bool) {
			t.Helper()

			// Dial the server.
			protocol, address := osm.ProtocolAndAddress(address)
			conn, err := net.Dial(protocol, address)
			if err == nil {
				defer conn.Close()
			}

			// If noWrite is true, return without writing anything.
			if noWrite {
				return // Do not write anything.
			}

			// Write a ping request then close the connection.
			amino.MarshalAnySizedWriter(conn, &r.PingRequest{})
		}

		unixSocket := testUnixSocket(t)

		rss := newRemoteSignerServer(t, unixSocket, nil)
		require.NoError(t, rss.Start())
		require.NotNil(t, rss)

		for i := 0; i < 100; i++ {
			newReadWriteErrorRemoteSignerClient(t, unixSocket, false)
			newReadWriteErrorRemoteSignerClient(t, unixSocket, true)
		}

		require.True(t, rss.IsRunning())
		assert.NoError(t, rss.Stop())
	})

	t.Run("tcp configuration succeeded", func(t *testing.T) {
		t.Parallel()

		// Client succeeded authenticating server.
		clientPrivKey := ed25519.GenPrivKey()
		rss, err := NewRemoteSignerServer(
			types.NewMockSigner(),
			[]string{tcpLocalhost + ":0"},
			log.NewNoopLogger(),
			WithAuthorizedKeys([]ed25519.PubKeyEd25519{clientPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		serverPort := rss.listeners[0].Addr().(*net.TCPAddr).Port
		rsc, err := c.NewRemoteSignerClient(
			fmt.Sprintf("%s:%d", tcpLocalhost, serverPort),
			log.NewNoopLogger(),
			c.WithClientPrivKey(clientPrivKey),
		)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.NoError(t, rsc.Ping())
		rss.Stop()
		rsc.Close()

		// Server succeeded authenticating client.
		serverPrivKey := ed25519.GenPrivKey()
		rss, err = NewRemoteSignerServer(
			types.NewMockSigner(),
			[]string{tcpLocalhost + ":0"},
			log.NewNoopLogger(),
			WithServerPrivKey(serverPrivKey),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		serverPort = rss.listeners[0].Addr().(*net.TCPAddr).Port
		rsc, err = c.NewRemoteSignerClient(
			fmt.Sprintf("%s:%d", tcpLocalhost, serverPort),
			log.NewNoopLogger(),
			c.WithAuthorizedKeys([]ed25519.PubKeyEd25519{serverPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.NoError(t, rsc.Ping())
		rss.Stop()
		rsc.Close()
	})

	t.Run("tcp configuration failed", func(t *testing.T) {
		t.Parallel()

		// Client fails authenticating server.
		rss, err := NewRemoteSignerServer(
			types.NewMockSigner(),
			[]string{tcpLocalhost + ":0"},
			log.NewNoopLogger(),
			WithAuthorizedKeys([]ed25519.PubKeyEd25519{ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		serverPort := rss.listeners[0].Addr().(*net.TCPAddr).Port
		rsc, err := c.NewRemoteSignerClient(
			fmt.Sprintf("%s:%d", tcpLocalhost, serverPort),
			log.NewNoopLogger(),
			c.WithAuthorizedKeys([]ed25519.PubKeyEd25519{ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.ErrorIs(t, rsc.Ping(), r.ErrUnauthorizedPubKey)
		rss.Stop()
		rsc.Close()
	})
}

package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	r "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote"
	s "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/server"
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
	tcpTimeouts    = 100 * time.Millisecond
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

func newRemoteSignerClient(t *testing.T, address string) (*RemoteSignerClient, context.CancelFunc) {
	t.Helper()

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)

	rsc, err := NewRemoteSignerClient(
		ctx,
		address,
		log.NewNoopLogger(),
		WithDialMaxRetries(3),
		WithDialTimeout(tcpTimeouts),
		WithDialRetryInterval(tcpTimeouts),
		WithRequestTimeout(tcpTimeouts),
		WithKeepAlivePeriod(tcpTimeouts),
	)
	require.NoError(t, err)

	return rsc, cancelFn
}

func newRemoteSignerServer(t *testing.T, address string, signer types.Signer) *s.RemoteSignerServer {
	t.Helper()

	if signer == nil {
		signer = types.NewMockSigner()
	}

	rss, err := s.NewRemoteSignerServer(
		signer,
		address,
		log.NewNoopLogger(),
		s.WithKeepAlivePeriod(tcpTimeouts),
		s.WithResponseTimeout(tcpTimeouts),
	)
	require.NoError(t, err)

	return rss
}

func TestCloseState(t *testing.T) {
	t.Parallel()

	t.Run("basic", func(t *testing.T) {
		t.Parallel()

		unixSocket := testUnixSocket(t)

		// Init a remote signer server.
		rss := newRemoteSignerServer(t, unixSocket, types.NewMockSigner())
		require.NoError(t, rss.Start())
		defer rss.Stop()

		// Init a new remote signer client.
		rsc, cancelFn := newRemoteSignerClient(t, unixSocket)
		defer cancelFn()
		require.NoError(t, rsc.ctx.Err())

		// Close it.
		require.NoError(t, rsc.Close())
		require.Error(t, rsc.ctx.Err())

		// Try to close it again.
		require.Error(t, rsc.Close())
	})

	t.Run("connection cleanup", func(t *testing.T) {
		t.Parallel()

		unixSocket := testUnixSocket(t)

		// Init a new remote signer server and client.
		rss := newRemoteSignerServer(t, unixSocket, nil)
		require.NoError(t, rss.Start())
		defer rss.Stop()
		rsc, cancelFn := newRemoteSignerClient(t, unixSocket)
		defer cancelFn()

		// Trigger a connection.
		_, err := rsc.Sign([]byte("test"))
		require.NoError(t, err)
		rsc.connLock.RLock()
		require.NotNil(t, rsc.conn)
		rsc.connLock.RUnlock()
		require.NoError(t, rsc.ctx.Err())

		// Close and check if connexion is closed.
		require.NoError(t, rsc.Close())
		require.Error(t, rsc.ctx.Err())
		rsc.connLock.RLock()
		assert.Nil(t, rsc.conn)
		rsc.connLock.RUnlock()
	})
}

func TestClientRequest(t *testing.T) {
	t.Parallel()

	// Faulty remote signer error.
	errFaultyServer := fmt.Errorf("faulty server")

	// Faulty remote signer server that returns an invalid message type or an error on signing.
	newFaultyRemoteSignerServer := func(t *testing.T, address string, erroring bool, wg *sync.WaitGroup) <-chan func() {
		t.Helper()

		// Create a listener.
		protocol, address := osm.ProtocolAndAddress(address)
		listener, err := net.Listen(protocol, address)
		require.NoError(t, err)

		// Create a channel to return the connection Close function.
		closer := make(chan func())

		// Generate an identity key pair.
		idPrivKey := ed25519.GenPrivKey()

		wg.Add(1)
		go func() {
			// Cleanup before returning.
			defer func() {
				listener.Close()
				wg.Done()
			}()

			// Accept the connection from the client.
			conn, err := listener.Accept()
			require.NoError(t, err)

			var (
				invalid  = []byte("invalid message type")
				firstReq = true
			)

			// Respond to client requests with an invalid message type.
			for {
				var (
					request  r.RemoteSignerMessage
					response r.RemoteSignerMessage
				)

				// Receive the request from the client and unmarshal it using amino.
				if _, err := amino.UnmarshalSizedReader(conn, &request, r.MaxMessageSize); err != nil {
					return // Connection closed.
				}

				switch request.(type) {
				case *r.PubKeyRequest:
					response = &r.PubKeyResponse{PubKey: idPrivKey.PubKey()}
				case *r.SignRequest:
					if erroring {
						response = &r.SignResponse{Signature: nil, Error: &r.RemoteSignerError{Err: errFaultyServer.Error()}}
					} else {
						response = invalid
					}
				default:
					response = invalid
				}

				// Send the response to the client.
				if _, err := amino.MarshalAnySizedWriter(conn, response); err != nil {
					return // Connection closed.
				}

				// Send the closer after the first request (client init).
				if firstReq {
					firstReq = false
					closer <- func() { conn.Close() }
				}
			}
		}()

		return closer
	}

	t.Run("PubKey request", func(t *testing.T) {
		t.Parallel()

		var (
			unixSocket = testUnixSocket(t)
			signer     = types.NewMockSigner()
		)

		// Init a new remote signer server and client.
		rss := newRemoteSignerServer(t, unixSocket, signer)
		require.NoError(t, rss.Start())
		rsc, cancelFn := newRemoteSignerClient(t, unixSocket)
		defer cancelFn()

		// Test a valid PubKey request.
		remotePK := rsc.PubKey()
		require.NotNil(t, remotePK)
		localPK := signer.PubKey()
		require.NotNil(t, localPK)
		require.Equal(t, localPK, remotePK)
		rss.Stop()
		rsc.Close()
	})

	t.Run("Sign request", func(t *testing.T) {
		t.Parallel()

		var (
			unixSocket = testUnixSocket(t)
			signer     = types.NewMockSigner()
			wg         = new(sync.WaitGroup)
		)

		// Init a new remote signer server and client.
		rss := newRemoteSignerServer(t, unixSocket, signer)
		require.NoError(t, rss.Start())
		rsc, cancelFn := newRemoteSignerClient(t, unixSocket)
		defer cancelFn()

		// Test a valid Sign request.
		remoteSignature, err := rsc.Sign([]byte("sign bytes"))
		require.NotNil(t, remoteSignature)
		require.NoError(t, err)
		localSignature, err := signer.Sign([]byte("sign bytes"))
		require.NotNil(t, localSignature)
		require.NoError(t, err)
		require.Equal(t, localSignature, remoteSignature)
		rss.Stop()
		rsc.Close()

		// Init a erroring remote signer server and a regular client.
		unixSocket = testUnixSocket(t)
		chanCloser := newFaultyRemoteSignerServer(t, unixSocket, true, wg)
		rsc, cancelFn = newRemoteSignerClient(t, unixSocket)
		defer cancelFn()
		require.NoError(t, rsc.ensureConnection())
		closer := <-chanCloser

		// Test an erroring Sign request.
		remoteSignature, err = rsc.Sign([]byte("sign bytes"))
		require.Nil(t, remoteSignature)
		require.ErrorIs(t, err, ErrResponseContainsError)
		require.Contains(t, err.Error(), errFaultyServer.Error())
		closer()
		rsc.Close()

		// Init a invalid remote signer server and a regular client.
		unixSocket = testUnixSocket(t)
		chanCloser = newFaultyRemoteSignerServer(t, unixSocket, false, wg)
		rsc, cancelFn = newRemoteSignerClient(t, unixSocket)
		defer cancelFn()
		require.NoError(t, rsc.ensureConnection())
		closer = <-chanCloser

		// Test an invalid Sign request.
		remoteSignature, err = rsc.Sign([]byte("sign bytes"))
		require.Nil(t, remoteSignature)
		require.ErrorIs(t, err, ErrInvalidResponseType)
		closer()

		// Test a failed Sign request.
		rsc.Close()
		remoteSignature, err = rsc.Sign([]byte("sign bytes"))
		require.Nil(t, remoteSignature)
		assert.ErrorIs(t, err, ErrSendingRequestFailed)

		wg.Wait()
	})

	t.Run("String method and cache", func(t *testing.T) {
		t.Parallel()

		var (
			unixSocket = testUnixSocket(t)
			signer     = types.NewMockSigner()
		)

		// Init a new remote signer server and client.
		rss := newRemoteSignerServer(t, unixSocket, signer)
		require.NoError(t, rss.Start())
		rsc, cancelFn := newRemoteSignerClient(t, unixSocket)
		defer cancelFn()

		// Check if the public key is cached.
		require.NotNil(t, rsc.cachedPubKey)

		// Check if the String method returns the address.
		pk := signer.PubKey()
		require.NotNil(t, pk)
		require.Contains(t, rsc.String(), pk.Address().String())
		rss.Stop()
		rsc.Close()
	})
}

func TestClientConnection(t *testing.T) {
	t.Parallel()

	t.Run("tcp configuration succeeded", func(t *testing.T) {
		t.Parallel()

		// Server succeeded authenticating client.
		serverPrivKey := ed25519.GenPrivKey()
		rss, err := s.NewRemoteSignerServer(
			types.NewMockSigner(),
			tcpLocalhost+":0",
			log.NewNoopLogger(),
			s.WithServerPrivKey(serverPrivKey),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		serverPort := rss.ListenAddress(t).(*net.TCPAddr).Port

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()
		rsc, err := NewRemoteSignerClient(
			ctx,
			fmt.Sprintf("%s:%d", tcpLocalhost, serverPort),
			log.NewNoopLogger(),
			WithAuthorizedKeys([]ed25519.PubKeyEd25519{serverPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.NoError(t, rsc.ensureConnection())
		rss.Stop()
		rsc.Close()

		// Client succeeded authenticating server.
		clientPrivKey := ed25519.GenPrivKey()
		rss, err = s.NewRemoteSignerServer(
			types.NewMockSigner(),
			tcpLocalhost+":0",
			log.NewNoopLogger(),
			s.WithAuthorizedKeys([]ed25519.PubKeyEd25519{clientPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		serverPort = rss.ListenAddress(t).(*net.TCPAddr).Port

		ctx, cancelFn = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()
		rsc, err = NewRemoteSignerClient(
			ctx,
			fmt.Sprintf("%s:%d", tcpLocalhost, serverPort),
			log.NewNoopLogger(),
			WithClientPrivKey(clientPrivKey),
		)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		assert.NoError(t, rsc.ensureConnection())
		rss.Stop()
		rsc.Close()
	})

	t.Run("tcp configuration failed", func(t *testing.T) {
		t.Parallel()

		// Server fails authenticating client.
		rss := newRemoteSignerServer(t, tcpLocalhost+":0", types.NewMockSigner())
		require.NoError(t, rss.Start())
		serverPort := rss.ListenAddress(t).(*net.TCPAddr).Port

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()
		rsc, err := NewRemoteSignerClient(
			ctx,
			fmt.Sprintf("%s:%d", tcpLocalhost, serverPort),
			log.NewNoopLogger(),
			WithAuthorizedKeys([]ed25519.PubKeyEd25519{ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.Nil(t, rsc)
		require.ErrorIs(t, err, r.ErrUnauthorizedPubKey)
		rss.Stop()

		// Check if the configuration fail with a nil connection.
		conn, err := r.ConfigureTCPConnection(nil, ed25519.PrivKeyEd25519{}, nil, r.TCPConnConfig{})
		require.Nil(t, conn)
		assert.ErrorIs(t, err, r.ErrNilConn)
	})
}

package client

import (
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
	"github.com/stretchr/testify/require"
)

const unixSocketPath = "/tmp/test_tm2_remote_signer"

func testUnixSocket(t *testing.T) string {
	t.Helper()

	return fmt.Sprintf("unix://%s/%s.sock", unixSocketPath, xid.New().String())
}

func newRemoteSignerClient(t *testing.T, address string) *RemoteSignerClient {
	t.Helper()

	rsc, _ := NewRemoteSignerClient(
		address,
		log.NewNoopLogger(),
	)

	return rsc
}

func newRemoteSignerServer(t *testing.T, address string, signer types.Signer) *s.RemoteSignerServer {
	t.Helper()

	if signer == nil {
		signer = types.NewMockSigner()
	}

	rss, _ := s.NewRemoteSignerServer(
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
		rsc := newRemoteSignerClient(t, testUnixSocket(t))
		require.False(t, rsc.isClosed())

		// Close it.
		require.NoError(t, rsc.Close())
		require.True(t, rsc.isClosed())

		// Try to close it again.
		require.Error(t, rsc.Close())
		require.True(t, rsc.isClosed())
	})

	t.Run("connection cleanup", func(t *testing.T) {
		t.Parallel()

		unixSocket := testUnixSocket(t)

		// Init a new remote signer server and client.
		rss := newRemoteSignerServer(t, unixSocket, nil)
		defer rss.Stop()
		require.NoError(t, rss.Start())
		rsc := newRemoteSignerClient(t, unixSocket)

		// Trigger a connection.
		require.NoError(t, rsc.Ping())
		rsc.connLock.RLock()
		require.NotNil(t, rsc.conn)
		rsc.connLock.RUnlock()
		require.False(t, rsc.isClosed())

		// Close and check if connexion is closed.
		require.NoError(t, rsc.Close())
		require.True(t, rsc.isClosed())
		rsc.connLock.RLock()
		require.Nil(t, rsc.conn)
		rsc.connLock.RUnlock()
	})
}

func TestClientRequest(t *testing.T) {
	t.Parallel()

	// Create a wait group for the faulty server goroutines.
	wg := new(sync.WaitGroup)

	// Create a directory for the unix socket.
	os.MkdirAll(unixSocketPath, 0o755)

	// Remove the directory after the test.
	t.Cleanup(func() {
		wg.Wait()
		os.Remove(unixSocketPath)
	})

	// Faulty remote signer error.
	errFaultyServer := fmt.Errorf("faulty server")

	// Faulty remote signer server.
	newFaultyRemoteSignerServer := func(t *testing.T, address string, erroring bool, wg *sync.WaitGroup) <-chan func() {
		t.Helper()

		// Create a listener.
		protocol, address := osm.ProtocolAndAddress(address)
		listener, err := net.Listen(protocol, address)
		require.NoError(t, err)

		// Create a channel to return the connection Close function.
		closer := make(chan func())

		wg.Add(1)
		go func() {
			// Cleanup before returning.
			defer func() {
				listener.Close()
				wg.Done()
			}()

			// Accept a connection and send its Close function.
			conn, err := listener.Accept()
			require.NoError(t, err)
			closer <- func() { conn.Close() }

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

				// Always return an error.
				if erroring {
					switch request.(type) {
					case *r.PubKeyRequest:
						response = &r.PubKeyResponse{PubKey: nil, Error: &r.RemoteSignerError{Err: errFaultyServer.Error()}}
					case *r.SignRequest:
						response = &r.SignResponse{Signature: nil, Error: &r.RemoteSignerError{Err: errFaultyServer.Error()}}
					}
				} else {
					// Always return an invalid message type.
					response = []byte("invalid message type")
				}

				// Send the response to the client.
				if _, err := amino.MarshalAnySizedWriter(conn, response); err != nil {
					return // Connection closed.
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
		rsc := newRemoteSignerClient(t, unixSocket)

		// Test a valid PubKey request.
		remotePK, err := rsc.PubKey()
		require.NotNil(t, remotePK)
		require.NoError(t, err)
		localPK, err := signer.PubKey()
		require.NotNil(t, localPK)
		require.NoError(t, err)
		require.Equal(t, localPK, remotePK)
		rss.Stop()
		rsc.Close()

		// Init a erroring remote signer server and a regular client.
		unixSocket = testUnixSocket(t)
		chanCloser := newFaultyRemoteSignerServer(t, unixSocket, true, wg)
		rsc = newRemoteSignerClient(t, unixSocket)
		require.NoError(t, rsc.ensureConnection())
		closer := <-chanCloser

		// Test an erroring PubKey request.
		remotePK, err = rsc.PubKey()
		require.Nil(t, remotePK)
		require.ErrorIs(t, err, ErrResponseContainsError)
		require.Contains(t, err.Error(), errFaultyServer.Error())
		closer()
		rsc.Close()

		// Init a invalid remote signer server and a regular client.
		unixSocket = testUnixSocket(t)
		chanCloser = newFaultyRemoteSignerServer(t, unixSocket, false, wg)
		rsc = newRemoteSignerClient(t, unixSocket)
		require.NoError(t, rsc.ensureConnection())
		closer = <-chanCloser

		// Test an invalid PubKey request.
		remotePK, err = rsc.PubKey()
		require.Nil(t, remotePK)
		require.ErrorIs(t, err, ErrInvalidResponseType)
		closer()

		// Test a failed PubKey request.
		rsc.Close()
		remotePK, err = rsc.PubKey()
		require.Nil(t, remotePK)
		require.ErrorIs(t, err, ErrSendingRequestFailed)
	})

	t.Run("Sign request", func(t *testing.T) {
		t.Parallel()

		var (
			unixSocket = testUnixSocket(t)
			signer     = types.NewMockSigner()
		)

		// Init a new remote signer server and client.
		rss := newRemoteSignerServer(t, unixSocket, signer)
		require.NoError(t, rss.Start())
		rsc := newRemoteSignerClient(t, unixSocket)

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
		rsc = newRemoteSignerClient(t, unixSocket)
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
		rsc = newRemoteSignerClient(t, unixSocket)
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
		require.ErrorIs(t, err, ErrSendingRequestFailed)
	})

	t.Run("Ping request", func(t *testing.T) {
		t.Parallel()

		var (
			unixSocket = testUnixSocket(t)
			signer     = types.NewMockSigner()
		)

		// Init a new remote signer server and client.
		rss := newRemoteSignerServer(t, unixSocket, signer)
		require.NoError(t, rss.Start())
		rsc := newRemoteSignerClient(t, unixSocket)

		// Test a valid Ping request.
		err := rsc.Ping()
		require.NoError(t, err)
		rss.Stop()
		rsc.Close()

		// Init a erroring remote signer server and a regular client.
		unixSocket = testUnixSocket(t)
		chanCloser := newFaultyRemoteSignerServer(t, unixSocket, false, wg)
		rsc = newRemoteSignerClient(t, unixSocket)
		require.NoError(t, rsc.ensureConnection())
		closer := <-chanCloser

		// Test an invalid Ping request.
		err = rsc.Ping()
		require.ErrorIs(t, err, ErrInvalidResponseType)
		closer()

		// Test a failed Ping request.
		rsc.Close()
		err = rsc.Ping()
		require.ErrorIs(t, err, ErrSendingRequestFailed)
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
		rsc := newRemoteSignerClient(t, unixSocket)

		// Check if the address is not cached
		require.Empty(t, rsc.addrCache)

		// Check if the String method returns the address.
		pk, err := signer.PubKey()
		require.NotNil(t, pk)
		require.NoError(t, err)
		require.Contains(t, rsc.String(), pk.Address().String())
		require.Contains(t, rsc.addrCache, pk.Address().String())
		rss.Stop()
		rsc.Close()
	})
}

func TestClientConnection(t *testing.T) {
	t.Parallel()

	// Create a wait group for the faulty server goroutines.
	wg := new(sync.WaitGroup)

	// Create a directory for the unix socket.
	os.MkdirAll(unixSocketPath, 0o755)

	// Remove the directory after the test.
	t.Cleanup(func() {
		wg.Wait()
		os.Remove(unixSocketPath)
	})

	// Server that fails on read.
	newReadWriteErrorRemoteSignerServer := func(t *testing.T, address string, wg *sync.WaitGroup) {
		t.Helper()

		// Create a listener.
		protocol, address := osm.ProtocolAndAddress(address)
		listener, err := net.Listen(protocol, address)
		require.NoError(t, err)

		wg.Add(1)
		go func() {
			// Cleanup before returning.
			defer func() {
				listener.Close()
				wg.Done()
			}()

			conn, err := listener.Accept()
			require.NoError(t, err)
			conn.Close()
		}()
	}

	t.Run("conn closed during read/write", func(t *testing.T) {
		t.Parallel()

		// Repeat to catch both read and write errors.
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				unixSocket := testUnixSocket(t)

				// Init a new remote signer server and client.
				newReadWriteErrorRemoteSignerServer(t, unixSocket, wg)
				rsc, err := NewRemoteSignerClient(
					unixSocket,
					log.NewNoopLogger(),
					WithDialMaxRetries(3),
					WithDialRetryInterval(time.Microsecond),
					WithRequestTimeout(time.Millisecond),
					WithDialTimeout(time.Millisecond),
				)
				require.NotNil(t, rsc)
				require.NoError(t, err)
				require.ErrorIs(t, rsc.Ping(), ErrMaxRetriesExceeded)
			}()
		}
	})

	t.Run("force close whiie trying to dial", func(t *testing.T) {
		t.Parallel()

		unixSocket := testUnixSocket(t)

		// Init a new remote signer server and client.
		rsc, err := NewRemoteSignerClient(
			unixSocket,
			log.NewNoopLogger(),
			WithDialRetryInterval(time.Microsecond),
			WithRequestTimeout(time.Microsecond),
			WithDialTimeout(time.Microsecond),
		)
		require.NotNil(t, rsc)
		require.NoError(t, err)

		// Close the client while it is trying to dial.
		wg.Add(1)
		go func() {
			defer wg.Done()
			require.NoError(t, rsc.ensureConnection())
		}()
		time.Sleep(10 * time.Millisecond)
		rsc.Close()
	})

	getFreePort := func(t *testing.T) int {
		t.Helper()

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		port := listener.Addr().(*net.TCPAddr).Port
		listener.Close()

		return port
	}

	t.Run("tcp configuration succeeded", func(t *testing.T) {
		t.Parallel()

		tcpAddr := fmt.Sprintf("tcp://127.0.0.1:%d", getFreePort(t))

		// Server succeeded authenticating client.
		serverPrivKey := ed25519.GenPrivKey()
		rss, err := s.NewRemoteSignerServer(
			types.NewMockSigner(),
			[]string{tcpAddr},
			log.NewNoopLogger(),
			s.WithServerPrivKey(serverPrivKey),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		rsc, err := NewRemoteSignerClient(
			tcpAddr,
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
			[]string{tcpAddr},
			log.NewNoopLogger(),
			s.WithAuthorizedKeys([]ed25519.PubKeyEd25519{clientPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		rsc, err = NewRemoteSignerClient(
			tcpAddr,
			log.NewNoopLogger(),
			WithClientPrivKey(clientPrivKey),
		)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.NoError(t, rsc.ensureConnection())
		rss.Stop()
		rsc.Close()
	})

	// Conn close during handshake.
	connCloseDuringHandshake := func(t *testing.T, address string, wg *sync.WaitGroup) {
		t.Helper()

		// Create a listener.
		protocol, address := osm.ProtocolAndAddress(address)
		listener, err := net.Listen(protocol, address)
		require.NoError(t, err)

		wg.Add(1)
		go func() {
			// Cleanup before returning.
			defer func() {
				listener.Close()
				wg.Done()
			}()

			conn, err := listener.Accept()
			require.NoError(t, err)

			// Read the first byte of the handshake.
			conn.Read(make([]byte, 1))
			conn.Close()
		}()
	}

	t.Run("tcp configuration failed", func(t *testing.T) {
		t.Parallel()

		tcpAddr := fmt.Sprintf("tcp://127.0.0.1:%d", getFreePort(t))

		// Server fails authenticating client.
		rss := newRemoteSignerServer(t, tcpAddr, types.NewMockSigner())
		require.NoError(t, rss.Start())
		rsc, err := NewRemoteSignerClient(
			tcpAddr,
			log.NewNoopLogger(),
			WithAuthorizedKeys([]ed25519.PubKeyEd25519{ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rsc)
		require.NoError(t, err)
		require.ErrorIs(t, rsc.ensureConnection(), r.ErrUnauthorizedPubKey)
		rss.Stop()
		rsc.Close()

		// Client fails authenticating server.
		rss, err = s.NewRemoteSignerServer(
			types.NewMockSigner(),
			[]string{tcpAddr},
			log.NewNoopLogger(),
			s.WithAuthorizedKeys([]ed25519.PubKeyEd25519{ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)}),
		)
		require.NotNil(t, rss)
		require.NoError(t, err)
		require.NoError(t, rss.Start())
		rsc = newRemoteSignerClient(t, tcpAddr)
		require.NotNil(t, rsc)
		require.NoError(t, rsc.ensureConnection())
		rss.Stop()
		rsc.Close()

		// Check if the configuration fail with a nil connection.
		conn, err := r.ConfigureTCPConnection(nil, ed25519.PrivKeyEd25519{}, nil, 0, 0)
		require.Nil(t, conn)
		require.ErrorIs(t, err, r.ErrNilConn)

		// Check if the configuration fail during the handshake.
		connCloseDuringHandshake(t, tcpAddr, wg)
		protocol, address := osm.ProtocolAndAddress(tcpAddr)
		tcpConn, err := net.Dial(protocol, address)
		require.NotNil(t, tcpConn)
		require.NoError(t, err)
		defer tcpConn.Close()
		conn, err = r.ConfigureTCPConnection(tcpConn.(*net.TCPConn), ed25519.GenPrivKey(), nil, 0, 0)
		require.Nil(t, conn)
		require.ErrorIs(t, err, r.ErrSecretConnFailed)
	})
}

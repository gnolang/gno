package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	r "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote"
	c "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote/client"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	p2pconn "github.com/gnolang/gno/tm2/pkg/p2p/conn"
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

func newRemoteSignerServer(t *testing.T, address string, signer types.Signer, opts ...Option) *RemoteSignerServer {
	t.Helper()

	if signer == nil {
		signer = types.NewMockSigner()
	}

	rss, err := NewRemoteSignerServer(
		signer,
		address,
		log.NewNoopLogger(),
		opts...,
	)
	require.NoError(t, err)

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

func TestServerSilentPeer(t *testing.T) {
	t.Parallel()

	// The server waits 10s for a handshake, the client 2s.
	const (
		responseTimeout = 5 * time.Second
		requestTimeout  = 2 * time.Second
	)

	clientPrivKey := ed25519.GenPrivKey()
	rss := newRemoteSignerServer(t, tcpLocalhost+":0", nil,
		WithResponseTimeout(responseTimeout),
		WithAuthorizedKeys([]ed25519.PubKeyEd25519{clientPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
	)
	require.NoError(t, rss.Start())
	defer rss.Stop()
	serverAddr := rss.ListenAddress(t).String()

	// A peer with no key connects and never answers the handshake.
	dialHalfway(t, serverAddr)

	// The authorized client must still get a signature, well before the
	// silent peer's handshake deadline.
	ctx, cancelFn := context.WithTimeout(context.Background(), requestTimeout)
	defer cancelFn()
	rsc, err := newTCPClient(ctx, serverAddr, clientPrivKey, requestTimeout)
	require.NoError(t, err)
	defer rsc.Close()
	_, err = rsc.Sign([]byte("test"))
	require.NoError(t, err)
}

// newTCPClient connects a client that makes one dial attempt and gives up on a handshake or a request after timeout.
func newTCPClient(
	ctx context.Context,
	address string,
	privKey ed25519.PrivKeyEd25519,
	timeout time.Duration,
) (*c.RemoteSignerClient, error) {
	return c.NewRemoteSignerClient(
		ctx,
		"tcp://"+address,
		log.NewNoopLogger(),
		c.WithClientPrivKey(privKey),
		c.WithRequestTimeout(timeout),
		c.WithDialMaxRetries(0),
	)
}

// dialHalfway opens a TCP connection to the server and returns it once the server
// has written the first byte of its handshake, a byte the returned reader still yields.
func dialHalfway(t *testing.T, address string) io.ReadWriteCloser {
	t.Helper()

	conn, err := net.Dial("tcp", address)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	require.NoError(t, conn.SetDeadline(time.Now().Add(5*time.Second)))

	first := make([]byte, 1)
	_, err = io.ReadFull(conn, first)
	require.NoError(t, err)

	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{io.MultiReader(bytes.NewReader(first), conn), conn, conn}
}

// finishAndRequest completes the handshake dialHalfway started, then sends a PubKey
// request and reads the response.
func finishAndRequest(
	rwc io.ReadWriteCloser,
	privKey ed25519.PrivKeyEd25519,
) (*p2pconn.SecretConnection, r.RemoteSignerMessage, error) {
	sconn, err := p2pconn.MakeSecretConnection(rwc, privKey)
	if err != nil {
		return nil, nil, err
	}
	if _, err := amino.MarshalAnySizedWriter(sconn, &r.PubKeyRequest{}); err != nil {
		return sconn, nil, err
	}

	var response r.RemoteSignerMessage
	_, err = amino.UnmarshalSizedReader(sconn, &response, r.MaxMessageSize)

	return sconn, response, err
}

// requireClosed fails unless err comes from the server closing the connection.
func requireClosed(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)
	var netErr net.Error
	require.False(t, errors.As(err, &netErr) && netErr.Timeout(), "timed out: %v", err)
}

func TestServerOneSessionAtATime(t *testing.T) {
	t.Parallel()

	clientPrivKey := ed25519.GenPrivKey()
	rss := newRemoteSignerServer(t, tcpLocalhost+":0", nil,
		WithAuthorizedKeys([]ed25519.PubKeyEd25519{clientPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
	)
	require.NoError(t, rss.Start())
	t.Cleanup(func() { rss.Stop() })
	serverAddr := rss.ListenAddress(t).String()

	// Of two handshakes running at once, the first to finish is served and the other closed.
	first := dialHalfway(t, serverAddr)
	second := dialHalfway(t, serverAddr)
	served, response, err := finishAndRequest(first, clientPrivKey)
	require.NoError(t, err)
	require.IsType(t, &r.PubKeyResponse{}, response)
	_, response, err = finishAndRequest(second, clientPrivKey)
	requireClosed(t, err)
	require.Nil(t, response)

	// A client arriving while a connection is served fails its handshake once its
	// request times out, rather than wait until its context ends.
	ctx, cancelFn := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFn()
	_, err = newTCPClient(ctx, serverAddr, clientPrivKey, 500*time.Millisecond)
	require.ErrorIs(t, err, r.ErrSecretConnFailed)
	require.NoError(t, ctx.Err())

	// A connection arriving as the served one closes is served once the server sees the close.
	require.NoError(t, served.Close())
	served, response, err = finishAndRequest(dialHalfway(t, serverAddr), clientPrivKey)
	require.NoError(t, err)
	require.IsType(t, &r.PubKeyResponse{}, response)

	// A handshake still running when the server stops is not served.
	require.NoError(t, served.Close())
	fourth := dialHalfway(t, serverAddr)
	require.NoError(t, rss.Stop())
	_, response, err = finishAndRequest(fourth, clientPrivKey)
	requireClosed(t, err)
	require.Nil(t, response)
}

// acceptCounter is a log handler counting the connections the server accepts and the responses it fails to send.
type acceptCounter struct {
	accepted, failedSends atomic.Int64
}

func (a *acceptCounter) Enabled(context.Context, slog.Level) bool { return true }
func (a *acceptCounter) WithAttrs([]slog.Attr) slog.Handler       { return a }
func (a *acceptCounter) WithGroup(string) slog.Handler            { return a }

func (a *acceptCounter) Handle(_ context.Context, record slog.Record) error {
	switch record.Message {
	case "Accepted new connection":
		a.accepted.Add(1)
	case "Failed to send response":
		a.failedSends.Add(1)
	}
	return nil
}

func TestServerBusyClientsDoNotReconnect(t *testing.T) {
	t.Parallel()

	clientPrivKey := ed25519.GenPrivKey()
	rss := newRemoteSignerServer(t, tcpLocalhost+":0", nil,
		WithAuthorizedKeys([]ed25519.PubKeyEd25519{clientPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
	)
	counter := &acceptCounter{}
	rss.logger = slog.New(counter)
	require.NoError(t, rss.Start())
	t.Cleanup(func() { rss.Stop() })
	serverAddr := rss.ListenAddress(t).String()

	_, _, err := finishAndRequest(dialHalfway(t, serverAddr), clientPrivKey)
	require.NoError(t, err)

	// Two clients arriving while a connection is served each time out once, rather
	// than close each other's connection and redial until their context ends.
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() {
			ctx, cancelFn := context.WithTimeout(context.Background(), time.Second)
			defer cancelFn()
			_, err := newTCPClient(ctx, serverAddr, clientPrivKey, 500*time.Millisecond)
			assert.ErrorIs(t, err, r.ErrSecretConnFailed)
			assert.NoError(t, ctx.Err())
		})
	}
	wg.Wait()

	// The served connection, then the one waiting for it; the other stays in the listener's backlog.
	require.EqualValues(t, 2, counter.accepted.Load())
}

func TestServerRestartParked(t *testing.T) {
	t.Parallel()

	clientPrivKey := ed25519.GenPrivKey()
	rss := newRemoteSignerServer(t, tcpLocalhost+":0", nil,
		WithAuthorizedKeys([]ed25519.PubKeyEd25519{clientPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
	)
	counter := &acceptCounter{}
	rss.logger = slog.New(counter)
	require.NoError(t, rss.Start())
	t.Cleanup(func() { rss.Stop() })
	serverAddr := rss.ListenAddress(t).String()

	// A connection accepted while another is served waits for it, then the server restarts.
	_, _, err := finishAndRequest(dialHalfway(t, serverAddr), clientPrivKey)
	require.NoError(t, err)
	parked, err := net.Dial("tcp", serverAddr)
	require.NoError(t, err)
	t.Cleanup(func() { parked.Close() })
	require.NoError(t, parked.SetDeadline(time.Now().Add(5*time.Second)))
	require.Eventually(t, func() bool { return counter.accepted.Load() == 2 }, time.Second, time.Millisecond)
	require.NoError(t, rss.Stop())
	require.NoError(t, rss.Start())

	// The restarted server does not serve a connection the stopped one accepted.
	_, response, err := finishAndRequest(parked, clientPrivKey)
	requireClosed(t, err)
	require.Nil(t, response)
}

// blockingSigner is a signer whose Sign reports it started, then returns once release is closed.
type blockingSigner struct {
	types.Signer
	signing, release chan struct{}
}

func (s *blockingSigner) Sign(signBytes []byte) ([]byte, error) {
	close(s.signing)
	<-s.release
	return s.Signer.Sign(signBytes)
}

func TestServerRestartMidSign(t *testing.T) {
	t.Parallel()

	clientPrivKey := ed25519.GenPrivKey()
	signer := &blockingSigner{types.NewMockSigner(), make(chan struct{}), make(chan struct{})}
	rss := newRemoteSignerServer(t, tcpLocalhost+":0", signer,
		WithAuthorizedKeys([]ed25519.PubKeyEd25519{clientPrivKey.PubKey().(ed25519.PubKeyEd25519)}),
	)
	counter := &acceptCounter{}
	rss.logger = slog.New(counter)
	require.NoError(t, rss.Start())
	t.Cleanup(func() { rss.Stop() })

	// The server restarts while a handler signs.
	signing, _, err := finishAndRequest(dialHalfway(t, rss.ListenAddress(t).String()), clientPrivKey)
	require.NoError(t, err)
	_, err = amino.MarshalAnySizedWriter(signing, &r.SignRequest{SignBytes: []byte("test")})
	require.NoError(t, err)
	<-signer.signing
	require.NoError(t, rss.Stop())
	require.NoError(t, rss.Start())

	// The handler returning does not close the connection the restarted server serves.
	served, _, err := finishAndRequest(dialHalfway(t, rss.ListenAddress(t).String()), clientPrivKey)
	require.NoError(t, err)
	close(signer.release)
	require.Eventually(t, func() bool { return counter.failedSends.Load() == 1 }, time.Second, time.Millisecond)
	_, err = amino.MarshalAnySizedWriter(served, &r.PubKeyRequest{})
	require.NoError(t, err)
	var response r.RemoteSignerMessage
	_, err = amino.UnmarshalSizedReader(served, &response, r.MaxMessageSize)
	require.NoError(t, err)
	require.IsType(t, &r.PubKeyResponse{}, response)
}

func TestServerHandshakeDeadline(t *testing.T) {
	t.Parallel()

	// With no response timeout, the handshake still has a deadline.
	rss := newRemoteSignerServer(t, tcpLocalhost+":0", nil, WithResponseTimeout(0))
	require.NoError(t, rss.Start())
	defer rss.Stop()

	// The server closes a peer that never answers the handshake, rather than wait for it.
	start := time.Now()
	silent, err := net.Dial("tcp", rss.ListenAddress(t).String())
	require.NoError(t, err)
	defer silent.Close()
	require.NoError(t, silent.SetReadDeadline(start.Add(DefaultResponseTimeout*2+5*time.Second)))
	_, err = io.Copy(io.Discard, silent)
	var netErr net.Error
	require.False(t, errors.As(err, &netErr) && netErr.Timeout())
	require.GreaterOrEqual(t, time.Since(start), DefaultResponseTimeout*2)
}

// failingListener is a listener whose Accept always fails, as it does once the process has
// no descriptor left.
type failingListener struct {
	net.Listener
	accepts atomic.Int64
}

func (l *failingListener) Accept() (net.Conn, error) {
	l.accepts.Add(1)
	return nil, errors.New("too many open files")
}

func TestServerAcceptBackoff(t *testing.T) {
	t.Parallel()

	rss := newRemoteSignerServer(t, testUnixSocket(t), nil)
	rss.setRunning(true)

	tcpListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer tcpListener.Close()
	listener := &failingListener{Listener: tcpListener}

	go rss.serve(listener)
	time.Sleep(200 * time.Millisecond)
	rss.setRunning(false)

	// Waits of 5, 10, 20, 40 and 80ms fit about six attempts in 200ms.
	require.Less(t, listener.accepts.Load(), int64(20))
}

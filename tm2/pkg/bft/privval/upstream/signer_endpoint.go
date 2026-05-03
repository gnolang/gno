package upstream

// signer_endpoint.go: base endpoint type shared by the listener and dialer
// variants. Holds a single live conn, owns read/write deadlines and the
// length-prefixed proto framing.
//
// Direct port of cometbft/privval/signer_endpoint.go (CometBFT v0.39.1).
// The structure, method names, and lifecycle pattern are kept identical
// to ease audit comparison. Only the dependencies are swapped — slog
// instead of cometbft/libs/log, sync.Mutex instead of cmtsync.Mutex,
// our local protoio.go instead of cometbft/libs/protoio, and
// upstreampb.Message instead of privvalproto.Message.

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/gnolang/gno/tm2/pkg/service"
)

const defaultTimeoutReadWriteSeconds = 5

// signerEndpoint is the unexported base type embedded by SignerListenerEndpoint
// (and, eventually, a SignerDialerEndpoint). Holds the shared state for
// one privval connection.
type signerEndpoint struct {
	service.BaseService

	connMtx sync.Mutex
	conn    net.Conn

	// connGen counts each install of a new conn on this endpoint. Used by
	// SignerClient to detect reconnects and re-verify the signer's pubkey
	// against its cached identity (defense against tmkms-instance swap
	// during a connection drop).
	connGen atomic.Uint64

	timeoutReadWrite time.Duration
}

// Close drops the underlying conn. Idempotent.
func (se *signerEndpoint) Close() error {
	se.DropConnection()
	return nil
}

// IsConnected reports whether a live conn is held.
func (se *signerEndpoint) IsConnected() bool {
	se.connMtx.Lock()
	defer se.connMtx.Unlock()
	return se.isConnected()
}

// GetAvailableConnection retrieves a queued conn if one is ready, without
// blocking. Returns true if a conn was claimed.
func (se *signerEndpoint) GetAvailableConnection(connectionAvailableCh chan net.Conn) bool {
	se.connMtx.Lock()
	defer se.connMtx.Unlock()

	select {
	case se.conn = <-connectionAvailableCh:
		se.connGen.Add(1)
		return true
	default:
	}
	return false
}

// WaitConnection blocks for up to maxWait waiting for a queued conn,
// returning ErrConnectionTimeout if none arrives in time.
func (se *signerEndpoint) WaitConnection(connectionAvailableCh chan net.Conn, maxWait time.Duration) error {
	select {
	case conn := <-connectionAvailableCh:
		se.SetConnection(conn)
	case <-time.After(maxWait):
		return ErrConnectionTimeout
	}
	return nil
}

// SetConnection installs a new conn, replacing any previously held one.
func (se *signerEndpoint) SetConnection(newConnection net.Conn) {
	se.connMtx.Lock()
	defer se.connMtx.Unlock()
	se.conn = newConnection
	se.connGen.Add(1)
}

// ConnectionGeneration returns a counter that increments every time a
// new conn is installed. Used by SignerClient to spot a reconnect and
// re-verify the signer's identity before signing for it.
func (se *signerEndpoint) ConnectionGeneration() uint64 {
	return se.connGen.Load()
}

// DropConnection closes and clears the held conn. Idempotent.
func (se *signerEndpoint) DropConnection() {
	se.connMtx.Lock()
	defer se.connMtx.Unlock()
	se.dropConnection()
}

// ReadMessage reads one privval message from the held conn, applying the
// configured read deadline. On timeout, the conn is dropped so the next
// read attempt forces a reconnect.
func (se *signerEndpoint) ReadMessage() (msg upstreampb.Message, err error) {
	se.connMtx.Lock()
	defer se.connMtx.Unlock()

	if !se.isConnected() {
		return msg, fmt.Errorf("endpoint is not connected: %w", ErrNoConnection)
	}

	deadline := time.Now().Add(se.timeoutReadWrite)
	if err = se.conn.SetReadDeadline(deadline); err != nil {
		return
	}

	r := NewDelimitedReader(se.conn, MaxRemoteSignerMsgSize)
	if _, err = r.ReadMsg(&msg); err != nil {
		if _, ok := err.(timeoutError); ok {
			err = fmt.Errorf("%v: %w", err, ErrReadTimeout)
			se.Logger.Debug("dropping conn on read timeout")
			se.dropConnection()
		}
		return
	}
	return
}

// WriteMessage writes one privval message to the held conn, applying the
// configured write deadline. On timeout, the conn is dropped.
func (se *signerEndpoint) WriteMessage(msg upstreampb.Message) (err error) {
	se.connMtx.Lock()
	defer se.connMtx.Unlock()

	if !se.isConnected() {
		return fmt.Errorf("endpoint is not connected: %w", ErrNoConnection)
	}

	deadline := time.Now().Add(se.timeoutReadWrite)
	if err = se.conn.SetWriteDeadline(deadline); err != nil {
		return
	}

	w := NewDelimitedWriter(se.conn)
	if _, err = w.WriteMsg(&msg); err != nil {
		if _, ok := err.(timeoutError); ok {
			err = fmt.Errorf("%v: %w", err, ErrWriteTimeout)
			se.Logger.Debug("dropping conn on write timeout")
			se.dropConnection()
		}
		return
	}
	return
}

// isConnected (lowercase): caller must hold connMtx.
func (se *signerEndpoint) isConnected() bool {
	return se.conn != nil
}

// dropConnection (lowercase): caller must hold connMtx.
func (se *signerEndpoint) dropConnection() {
	if se.conn != nil {
		if err := se.conn.Close(); err != nil {
			se.Logger.Error("signerEndpoint: drop conn", "err", err)
		}
		se.conn = nil
	}
}

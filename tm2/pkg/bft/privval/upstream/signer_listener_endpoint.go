package upstream

// signer_listener_endpoint.go: validator-side listener for an external
// privval signer (tmkms / Horcrux).
//
// Direct port of cometbft/privval/signer_listener_endpoint.go (CometBFT
// v0.39.1). Listens for the external signer to dial in; holds the
// connection live with periodic pings; reconnects on drop.
//
// The validator uses this via SignerClient (see signer_client.go) which
// implements PrivValidator. Mirrors the upstream Tendermint convention
// where the validator is the LISTENER and the signer is the DIALER —
// the signer host needs no inbound network surface.

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/privval/upstream/upstreampb"
	"github.com/gnolang/gno/tm2/pkg/service"
)

// defaultTimeoutAcceptSeconds bounds how long acceptNewConnection blocks
// in net.Listener.Accept before returning an error. Mirrors CometBFT's
// constant of the same name in socket_listeners.go.
const defaultTimeoutAcceptSeconds = 3

// SignerListenerEndpointOption is an optional parameter passed to the
// constructor. Mirrors the cometbft pattern.
type SignerListenerEndpointOption func(*SignerListenerEndpoint)

// SignerListenerEndpointTimeoutReadWrite sets the read/write deadline on
// the held conn (default 5s). The ping interval is derived as (timeout * 2/3).
func SignerListenerEndpointTimeoutReadWrite(timeout time.Duration) SignerListenerEndpointOption {
	return func(sl *SignerListenerEndpoint) { sl.timeoutReadWrite = timeout }
}

// SignerListenerEndpoint listens on a net.Listener for an external signer
// to dial in. Once connected, the connection is kept alive by sending
// PingRequest every (timeoutReadWrite * 2/3) and dropping/reconnecting
// on read/write timeout.
type SignerListenerEndpoint struct {
	signerEndpoint

	listener              net.Listener
	connectRequestCh      chan struct{}
	connectionAvailableCh chan net.Conn

	timeoutAccept   time.Duration
	acceptFailCount atomic.Uint32
	pingTimer       *time.Ticker
	pingInterval    time.Duration

	// Serializes public method access (SendRequest, WaitForConnection,
	// OnStop). Distinct from connMtx (which guards the conn pointer).
	instanceMtx sync.Mutex
}

// NewSignerListenerEndpoint constructs a listener endpoint over the
// given net.Listener. The listener should already be wrapped with any
// SecretConnection / mutual-auth layer the caller wants — see
// socket_listener.go for the standard wrapper.
func NewSignerListenerEndpoint(
	logger *slog.Logger,
	listener net.Listener,
	options ...SignerListenerEndpointOption,
) *SignerListenerEndpoint {
	sl := &SignerListenerEndpoint{
		listener:      listener,
		timeoutAccept: defaultTimeoutAcceptSeconds * time.Second,
	}
	sl.BaseService = *service.NewBaseService(logger, "SignerListenerEndpoint", sl)
	sl.timeoutReadWrite = defaultTimeoutReadWriteSeconds * time.Second

	for _, opt := range options {
		opt(sl)
	}
	return sl
}

// OnStart implements service.Service.
func (sl *SignerListenerEndpoint) OnStart() error {
	sl.connectRequestCh = make(chan struct{}, 1)
	sl.connectionAvailableCh = make(chan net.Conn)

	// Ping interval is 2/3 of the read/write timeout, matching CometBFT.
	sl.pingInterval = time.Duration(sl.timeoutReadWrite.Milliseconds()*2/3) * time.Millisecond
	sl.pingTimer = time.NewTicker(sl.pingInterval)

	go sl.serviceLoop()
	go sl.pingLoop()

	// Trigger the first connect attempt immediately.
	sl.connectRequestCh <- struct{}{}
	return nil
}

// OnStop implements service.Service.
func (sl *SignerListenerEndpoint) OnStop() {
	sl.instanceMtx.Lock()
	defer sl.instanceMtx.Unlock()
	_ = sl.Close()

	if sl.listener != nil {
		if err := sl.listener.Close(); err != nil {
			sl.Logger.Error("SignerListenerEndpoint: closing listener", "err", err)
			sl.listener = nil
		}
	}
	if sl.pingTimer != nil {
		sl.pingTimer.Stop()
	}
}

// WaitForConnection blocks for up to maxWait waiting for a connected
// signer. Returns ErrConnectionTimeout if no signer connects in time.
//
// Validator startup typically calls this once before consensus begins,
// so the validator's identity (returned by SignerClient.GetPubKey()) is
// available before the first vote is needed.
func (sl *SignerListenerEndpoint) WaitForConnection(maxWait time.Duration) error {
	sl.instanceMtx.Lock()
	defer sl.instanceMtx.Unlock()
	return sl.ensureConnection(maxWait)
}

// SendRequest writes one privval message and reads the response. Used by
// SignerClient for PubKeyRequest / SignVoteRequest / etc.
//
// Resets the ping timer on success — successful traffic counts as a
// keepalive, no need to send an explicit ping right after.
func (sl *SignerListenerEndpoint) SendRequest(request *upstreampb.Message) (*upstreampb.Message, error) {
	sl.instanceMtx.Lock()
	defer sl.instanceMtx.Unlock()
	return sl.sendRequestLocked(request)
}

// Lock and Unlock expose the per-instance mutex so callers can bracket
// a multi-RPC sequence atomically (e.g., SignerClient does PubKeyRequest
// + SignVoteRequest under one lock so a reconnect can't sneak a different
// signer in between the identity check and the vote signing). Pair with
// SendRequestLocked.
func (sl *SignerListenerEndpoint) Lock()   { sl.instanceMtx.Lock() }
func (sl *SignerListenerEndpoint) Unlock() { sl.instanceMtx.Unlock() }

// SendRequestLocked is SendRequest without taking the instance mutex.
// Caller MUST hold the lock via Lock().
func (sl *SignerListenerEndpoint) SendRequestLocked(request *upstreampb.Message) (*upstreampb.Message, error) {
	return sl.sendRequestLocked(request)
}

func (sl *SignerListenerEndpoint) sendRequestLocked(request *upstreampb.Message) (*upstreampb.Message, error) {
	if err := sl.ensureConnection(sl.timeoutAccept); err != nil {
		return nil, err
	}
	if err := sl.WriteMessage(request); err != nil {
		return nil, err
	}
	res, err := sl.ReadMessage()
	if err != nil {
		return nil, err
	}
	sl.pingTimer.Reset(sl.pingInterval)
	return res, nil
}

func (sl *SignerListenerEndpoint) ensureConnection(maxWait time.Duration) error {
	if sl.IsConnected() {
		return nil
	}
	if sl.GetAvailableConnection(sl.connectionAvailableCh) {
		return nil
	}
	sl.Logger.Info("SignerListener: blocking for connection")
	sl.triggerConnect()
	return sl.WaitConnection(sl.connectionAvailableCh, maxWait)
}

func (sl *SignerListenerEndpoint) acceptNewConnection() (net.Conn, error) {
	if !sl.IsRunning() || sl.listener == nil {
		return nil, fmt.Errorf("endpoint is closing")
	}
	sl.Logger.Info("SignerListener: listening for new connection")
	conn, err := sl.listener.Accept()
	if err != nil {
		sl.acceptFailCount.Add(1)
		return nil, err
	}
	sl.acceptFailCount.Store(0)
	return conn, nil
}

func (sl *SignerListenerEndpoint) triggerConnect() {
	select {
	case sl.connectRequestCh <- struct{}{}:
	default:
	}
}

func (sl *SignerListenerEndpoint) triggerReconnect() {
	sl.DropConnection()
	sl.triggerConnect()
}

func (sl *SignerListenerEndpoint) serviceLoop() {
	for {
		select {
		case <-sl.connectRequestCh:
			// On start, the listen-timeout path can queue a duplicate
			// connect-request while the first request is still connecting.
			// Drop the duplicate.
			if sl.IsConnected() {
				sl.Logger.Debug("SignerListener: already connected, dropping listen request")
				continue
			}

			conn, err := sl.acceptNewConnection()
			if err != nil {
				sl.Logger.Error("SignerListener: accept failed",
					"err", err, "failures", sl.acceptFailCount.Load())
				sl.triggerConnect()
				continue
			}

			// Hand off the conn to whoever's waiting in ensureConnection.
			sl.Logger.Info("SignerListener: connected")
			select {
			case sl.connectionAvailableCh <- conn:
			case <-sl.Quit():
				return
			}

		case <-sl.Quit():
			return
		}
	}
}

func (sl *SignerListenerEndpoint) pingLoop() {
	for {
		select {
		case <-sl.pingTimer.C:
			_, err := sl.SendRequest(WrapMsg(&upstreampb.PingRequest{}))
			if err != nil {
				sl.Logger.Error("SignerListener: ping timeout, reconnecting")
				sl.triggerReconnect()
			}
		case <-sl.Quit():
			return
		}
	}
}

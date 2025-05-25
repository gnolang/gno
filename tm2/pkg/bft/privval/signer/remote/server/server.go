package server

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"go.uber.org/multierr"
)

// RemoteSignerServer provides a service that forwards requests to a types.Signer.
type RemoteSignerServer struct {
	// Required config.
	signer        types.Signer
	listenAddress string
	logger        *slog.Logger

	// Optional connection config.
	keepAlivePeriod time.Duration // If 0, keep alive is disabled.
	responseTimeout time.Duration // If 0, no timeout is set. Requests reception is not timed out.

	// Optional authentication config.
	serverPrivKey  ed25519.PrivKeyEd25519  // Default is a random key.
	authorizedKeys []ed25519.PubKeyEd25519 // If empty, all keys are authorized.

	// Internal.
	listener net.Listener
	conn     net.Conn
	lock     sync.RWMutex
	running  atomic.Bool
}

// IsRunning returns true if the server is running.
func (rss *RemoteSignerServer) IsRunning() bool {
	return rss.running.Load()
}

// setRunning sets the running state of the server and returns true if the state was changed.
func (rss *RemoteSignerServer) setRunning(running bool) (changed bool) {
	return rss.running.CompareAndSwap(!running, running)
}

// Start starts the remote signer server.
func (rss *RemoteSignerServer) Start() error {
	// Check if the server is already started and set the running state.
	if !rss.setRunning(true) {
		return ErrServerAlreadyStarted
	}

	// The protocol validity was already checked by the NewRemoteSignerServer function.
	protocol, address := osm.ProtocolAndAddress(rss.listenAddress)

	// Create a listener. If the listener creation fails, stop the server and return an error.
	listener, err := net.Listen(protocol, address)
	if err != nil {
		rss.Stop()
		return fmt.Errorf("%w for listener %s://%s: %w", ErrListenFailed, protocol, address, err)
	}

	rss.logger.Info("Server started")

	// Start listening for incoming connections.
	rss.setListener(listener)
	go rss.serve(listener)

	return nil
}

// Stop stops the remote signer server.
func (rss *RemoteSignerServer) Stop() error {
	// Check if the server is already stopped and set the running state.
	if !rss.setRunning(false) {
		return ErrServerAlreadyStopped
	}

	// Close the listener and conn if any.
	err := multierr.Combine(
		rss.setListener(nil),
		rss.setConnection(nil),
	)

	rss.logger.Info("Server stopped")

	return err
}

// ListenAddress returns the listen address of the server.
// NOTE: This method is only used for testing purposes.
func (rss *RemoteSignerServer) ListenAddress(t *testing.T) net.Addr {
	t.Helper() // Mark the function as a test helper.

	rss.lock.RLock()
	defer rss.lock.RUnlock()

	if rss.listener == nil {
		return nil
	}

	return rss.listener.Addr()
}

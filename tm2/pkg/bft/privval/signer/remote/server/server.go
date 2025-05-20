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
	listener  net.Listener
	conns     []net.Conn
	connsLock sync.RWMutex
	running   atomic.Bool
	wg        sync.WaitGroup // Listeners and connections goroutines will register in this.
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

	var err error

	// The protocol validity was already checked by the NewRemoteSignerServer function.
	protocol, address := osm.ProtocolAndAddress(rss.listenAddress)

	// Create a listener. If the listener creation fails, stop the server and return an error.
	rss.listener, err = net.Listen(protocol, address)
	if err != nil {
		rss.Stop()
		return fmt.Errorf("%w for listener %s://%s: %w", ErrListenFailed, protocol, address, err)
	}

	// The listener accepts connections in a separate goroutine which is added to the wait group.
	rss.wg.Add(1)
	go func() {
		defer rss.wg.Done()
		rss.listen()
	}()

	rss.logger.Info("Server started")

	return nil
}

// Stop stops the remote signer server.
func (rss *RemoteSignerServer) Stop() error {
	// Check if the server is already stopped and set the running state.
	if !rss.setRunning(false) {
		return ErrServerAlreadyStopped
	}

	// Close all listeners.
	err := rss.closeListener()

	// Close all connections.
	rss.closeConnections()

	// Wait for all listeners and connections goroutines to stop.
	rss.wg.Wait()

	rss.logger.Info("Server stopped")

	return err
}

// Wait waits for the remote signer server to stop.
func (rss *RemoteSignerServer) Wait() {
	rss.wg.Wait()
}

// ListenAddress returns the listen address of the server.
// NOTE: This method is only used for testing purposes.
func (rss *RemoteSignerServer) ListenAddress(t *testing.T) net.Addr {
	t.Helper() // Mark the function as a test helper.

	if rss.listener == nil {
		return nil
	}

	return rss.listener.Addr()
}

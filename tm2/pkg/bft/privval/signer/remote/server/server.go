package server

import (
	"fmt"
	"net"
	"testing"

	osm "github.com/gnolang/gno/tm2/pkg/os"
)

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

	// For each listen address, create a listener.
	for i := range rss.listenAddresses {
		// The protocol validity was already checked by the NewRemoteSignerServer function.
		protocol, address := osm.ProtocolAndAddress(rss.listenAddresses[i])

		// Create a listener. If the listener creation fails, stop the server and return an error.
		listener, err := net.Listen(protocol, address)
		if err != nil {
			rss.Stop()
			return fmt.Errorf("%w for listener %s://%s: %w", ErrListenFailed, protocol, address, err)
		}

		// Add the listener to the server.
		rss.listenersLock.Lock()
		rss.listeners[i] = listener
		rss.listenersLock.Unlock()

		// The listener accepts connections in a separate goroutine which is added to the wait group.
		rss.wg.Add(1)
		go func(listener net.Listener) {
			defer rss.wg.Done()
			rss.listen(listener)
		}(listener)
	}

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
	err := rss.closeListeners()

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

// ListenAddresses returns the listen addresses of the server.
// NOTE: This method is only used for testing purposes.
func (rss *RemoteSignerServer) ListenAddresses(t *testing.T) []net.Addr {
	t.Helper() // Mark the function as a test helper.

	// Get the addresses of all listeners.
	rss.listenersLock.RLock()
	addrs := make([]net.Addr, len(rss.listeners))
	for i, listener := range rss.listeners {
		addrs[i] = listener.Addr()
	}
	rss.listenersLock.RUnlock()

	return addrs
}

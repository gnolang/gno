package server

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// RemoteSignerServer provides a service that forwards requests to a types.Signer.
type RemoteSignerServer struct {
	// Required config.
	signer          types.Signer
	listenAddresses []string
	logger          *slog.Logger

	// Optional connection config.
	keepAlivePeriod time.Duration // If 0, keep alive is disabled.
	responseTimeout time.Duration // If 0, no timeout is set. Requests reception is not timed out.

	// Optional authentication config.
	serverPrivKey  ed25519.PrivKeyEd25519  // Default is a random key.
	authorizedKeys []ed25519.PubKeyEd25519 // If empty, all keys are authorized.

	// Internal.
	listeners     []net.Listener
	listenersLock sync.Mutex
	conns         []net.Conn
	connsLock     sync.RWMutex
	running       atomic.Bool
	wg            sync.WaitGroup // Listeners and connections goroutines will register in this.
}

// Default connection config.
const (
	DefaultKeepAlivePeriod = 2 * time.Second
	DefaultResponseTimeout = 3 * time.Second
)

// Option is a functional option type used for optional configuration.
type Option func(*RemoteSignerServer)

// WithKeepAlivePeriod sets the keep alive period for the TCP connection to the client.
// If set to 0, keep alive is disabled. The default is 2 seconds.
func WithKeepAlivePeriod(period time.Duration) Option {
	return func(rss *RemoteSignerServer) {
		rss.keepAlivePeriod = period
	}
}

// WithResponseTimeout sets the timeout for sending response to the client.
// If set to 0, no timeout is set. The default is 3 seconds.
func WithResponseTimeout(timeout time.Duration) Option {
	return func(rss *RemoteSignerServer) {
		rss.responseTimeout = timeout
	}
}

// WithServerPrivKey sets the private key used by the server to authenticate with the client.
// The default is a random key.
func WithServerPrivKey(privKey ed25519.PrivKeyEd25519) Option {
	return func(rss *RemoteSignerServer) {
		rss.serverPrivKey = privKey
	}
}

// WithAuthorizedKeys sets the list of authorized public keys that the server will accept.
// If empty (default), all keys are authorized.
func WithAuthorizedKeys(keys []ed25519.PubKeyEd25519) Option {
	return func(rss *RemoteSignerServer) {
		rss.authorizedKeys = keys
	}
}

// NewRemoteSignerServer creates a new RemoteSignerServer with the required server address and
// logger. The server can be further configured using functional options.
func NewRemoteSignerServer(
	signer types.Signer,
	listenAddresses []string,
	logger *slog.Logger,
	options ...Option,
) (*RemoteSignerServer, error) {
	// Instantiate a RemoteSignerServer with default options.
	rss := &RemoteSignerServer{
		signer:          signer,
		listenAddresses: listenAddresses,
		logger:          logger,
		keepAlivePeriod: DefaultKeepAlivePeriod,
		responseTimeout: DefaultResponseTimeout,
		serverPrivKey:   ed25519.GenPrivKey(),
		listeners:       make([]net.Listener, len(listenAddresses)),
	}

	// Check if signer is nil.
	if signer == nil {
		return nil, ErrNilSigner
	}

	// At least one listen address must be provided.
	if len(listenAddresses) == 0 {
		return nil, ErrNoListenAddressProvided
	}

	// Check the protocol of each listener address.
	for _, listenAddress := range listenAddresses {
		protocol, _ := osm.ProtocolAndAddress(listenAddress)
		if protocol != "tcp" && protocol != "unix" {
			return nil, fmt.Errorf(
				"%w for listener %s: expected (tcp|unix), got %s",
				ErrInvalidAddressProtocol,
				listenAddress,
				protocol,
			)
		}
	}

	// Check if logger is nil.
	if logger == nil {
		return nil, ErrNilLogger
	}

	// Apply all the functional options to configure the server.
	for _, option := range options {
		option(rss)
	}

	return rss, nil
}

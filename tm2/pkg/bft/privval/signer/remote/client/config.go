package client

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// RemoteSignerClient implements types.Signer by connecting to a RemoteSignerServer.
type RemoteSignerClient struct {
	// Required config.
	protocol string
	address  string
	logger   *slog.Logger

	// Optional connection config.
	dialMaxRetries    int // If -1, retry indefinitely.
	dialRetryInterval time.Duration
	dialTimeout       time.Duration // If 0, no timeout is set.
	keepAlivePeriod   time.Duration // If 0, keep alive is disabled.
	requestTimeout    time.Duration // If 0, no timeout is set.

	// Optional authentication config.
	clientPrivKey  ed25519.PrivKeyEd25519  // Default is a random key.
	authorizedKeys []ed25519.PubKeyEd25519 // If empty, all keys are authorized.

	// Internal.
	conn      net.Conn
	connLock  sync.RWMutex
	closed    atomic.Bool
	addrCache string
}

// Default connection config.
const (
	DefaultDialMaxRetries    = -1 // Retry indefinitely.
	DefaultDialRetryInterval = 5 * time.Second
	DefaultDialTimeout       = 5 * time.Second
	DefaultKeepAlivePeriod   = 2 * time.Second
	DefaultRequestTimeout    = 5 * time.Second
)

// ClientOption is a functional option type used for optional configuration.
type ClientOption func(*RemoteSignerClient)

// WithDialMaxRetries sets the maximum number of retries when dialing the server.
// If set to -1 (default), the client will retry indefinitely.
func WithDialMaxRetries(maxRetries int) ClientOption {
	return func(rsc *RemoteSignerClient) {
		rsc.dialMaxRetries = maxRetries
	}
}

// WithDialRetryInterval sets the interval between dial retries when connecting to the server.
// The default is 5 seconds.
func WithDialRetryInterval(interval time.Duration) ClientOption {
	return func(rsc *RemoteSignerClient) {
		rsc.dialRetryInterval = interval
	}
}

// WithDialTimeout sets the timeout for dialing the server.
// If set to 0, no timeout is set. The default is 5 seconds.
func WithDialTimeout(timeout time.Duration) ClientOption {
	return func(rsc *RemoteSignerClient) {
		rsc.dialTimeout = timeout
	}
}

// WithKeepAlivePeriod sets the keep alive period for the TCP connection to the server.
// If set to 0, keep alive is disabled. The default is 2 seconds.
func WithKeepAlivePeriod(period time.Duration) ClientOption {
	return func(rsc *RemoteSignerClient) {
		rsc.keepAlivePeriod = period
	}
}

// WithRequestTimeout sets the timeout for sending requests to the server.
// If set to 0, no timeout is set. The default is 5 seconds.
func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(rsc *RemoteSignerClient) {
		rsc.requestTimeout = timeout
	}
}

// WithClientPrivKey sets the private key used by the client to authenticate with the server.
// The default is a random key.
func WithClientPrivKey(privKey ed25519.PrivKeyEd25519) ClientOption {
	return func(rsc *RemoteSignerClient) {
		rsc.clientPrivKey = privKey
	}
}

// WithAuthorizedKeys sets the list of authorized public keys that the client will accept.
// If empty (default), all keys are authorized.
func WithAuthorizedKeys(keys []ed25519.PubKeyEd25519) ClientOption {
	return func(rsc *RemoteSignerClient) {
		rsc.authorizedKeys = keys
	}
}

// NewRemoteSignerClient creates a new RemoteSignerClient with the required server address and
// logger. The client can be further configured using functional options.
func NewRemoteSignerClient(
	serverAddress string,
	logger *slog.Logger,
	options ...ClientOption,
) (*RemoteSignerClient, error) {
	// Instantiate a RemoteSignerClient with default options.
	rsc := &RemoteSignerClient{
		logger:            logger,
		dialMaxRetries:    DefaultDialMaxRetries,
		dialRetryInterval: DefaultDialRetryInterval,
		dialTimeout:       DefaultDialTimeout,
		keepAlivePeriod:   DefaultKeepAlivePeriod,
		requestTimeout:    DefaultRequestTimeout,
		clientPrivKey:     ed25519.GenPrivKey(),
	}

	// Parse the server address.
	rsc.protocol, rsc.address = osm.ProtocolAndAddress(serverAddress)
	if rsc.protocol != "tcp" && rsc.protocol != "unix" {
		return nil, fmt.Errorf("%w: expected (tcp|unix), got %s", ErrInvalidAddressProtocol, rsc.protocol)
	}

	// Check if logger is nil.
	if logger == nil {
		return nil, ErrNilLogger
	}

	// Apply all the functional options to configure the client.
	for _, option := range options {
		option(rsc)
	}

	return rsc, nil
}

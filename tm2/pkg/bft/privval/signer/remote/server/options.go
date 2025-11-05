package server

import (
	"fmt"
	"log/slog"
	"time"

	r "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

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
	listenAddress string,
	logger *slog.Logger,
	options ...Option,
) (*RemoteSignerServer, error) {
	// Instantiate a RemoteSignerServer with default options.
	rss := &RemoteSignerServer{
		signer:          signer,
		listenAddress:   listenAddress,
		logger:          logger,
		keepAlivePeriod: DefaultKeepAlivePeriod,
		responseTimeout: DefaultResponseTimeout,
		serverPrivKey:   ed25519.GenPrivKey(),
	}

	// Check if signer is nil.
	if signer == nil {
		return nil, ErrNilSigner
	}

	// Check the protocol of the listener address.
	protocol, _ := osm.ProtocolAndAddress(listenAddress)
	if protocol != r.TCPProtocol && protocol != r.UDSProtocol {
		return nil, fmt.Errorf(
			"%w for listener %s: expected (tcp|unix), got %s",
			ErrInvalidAddressProtocol,
			listenAddress,
			protocol,
		)
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

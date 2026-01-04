package remote

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	p2pconn "github.com/gnolang/gno/tm2/pkg/p2p/conn"
)

// Constants for the connection protocols.
const (
	UDSProtocol = "unix"
	TCPProtocol = "tcp"
)

// Errors returned by the TCP connection configuration.
var (
	ErrUnauthorizedPubKey = errors.New("unauthorized remote public key")
	ErrSecretConnFailed   = errors.New("secret connection handshake failed")
	ErrNilConn            = errors.New("nil connection")
)

type TCPConnConfig struct {
	KeepAlivePeriod  time.Duration
	HandshakeTimeout time.Duration
}

// configureTCPConnection configures the linger and keep alive options for a TCP connection.
// It also secures the connection and mutually authenticates with the remote peer using the
// provided localPrivKey and a list of authorizedKeys.
func ConfigureTCPConnection(
	conn *net.TCPConn,
	localPrivKey ed25519.PrivKeyEd25519,
	authorizedKeys []ed25519.PubKeyEd25519,
	cfg TCPConnConfig,
) (net.Conn, error) {
	// Check if the connection is nil.
	if conn == nil {
		return nil, ErrNilConn
	}

	// Set the linger option to 0 to close the connection immediately.
	conn.SetLinger(0)

	// If KeepAlive duration is not 0, set the keep alive options.
	if cfg.KeepAlivePeriod != 0 {
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(cfg.KeepAlivePeriod)
	}

	// If HandshakeTimeout is not 0, set the deadline for the secret connection handshake.
	if cfg.HandshakeTimeout != 0 {
		conn.SetDeadline(time.Now().Add(cfg.HandshakeTimeout))
	}

	// Secure the TCP connection and authenticate the remote signer.
	sconn, err := p2pconn.MakeSecretConnection(conn, localPrivKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSecretConnFailed, err)
	}

	// Reset the deadline after the secret connection handshake.
	if cfg.HandshakeTimeout != 0 {
		conn.SetDeadline(time.Time{})
	}

	// Check if the public key of the remote peer is authorized.
	if err := checkAuthorizedKeys(sconn.RemotePubKey(), authorizedKeys); err != nil {
		return nil, err
	}

	return sconn, err
}

// checkAuthorizedKeys checks if the public key of the remote peer is authorized.
func checkAuthorizedKeys(remotePubKey ed25519.PubKeyEd25519, authorizedKeys []ed25519.PubKeyEd25519) error {
	// If the whitelist is empty, skip the check
	if len(authorizedKeys) == 0 {
		return nil
	}

	// Check if the public key of the remote peer is authorized
	if slices.ContainsFunc(authorizedKeys, func(authorizedKey ed25519.PubKeyEd25519) bool {
		return authorizedKey.Equals(remotePubKey)
	}) {
		return nil
	}

	return ErrUnauthorizedPubKey
}

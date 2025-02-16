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

// Errors returned by the TCP connection configuration.
var (
	ErrUnauthorizedPubKey = errors.New("unauthorized remote public key")
	ErrSecretConnFailed   = errors.New("secret connection handshake failed")
	ErrTCPConfigFailed    = errors.New("TCP connection configuration failed")
)

// configureTCPConnection configures the linger and keep alive options for a TCP connection.
// It also secures the connection and mutually authenticates with the remote peer using the
// provided localPrivKey and a list of authorizedKeys.
func ConfigureTCPConnection(
	conn *net.TCPConn,
	localPrivKey ed25519.PrivKeyEd25519,
	authorizedKeys []ed25519.PubKeyEd25519,
	keepAlivePeriod time.Duration,
	handshakeTimeout time.Duration,
) (net.Conn, error) {
	// Set the linger option to 0 to close the connection immediately.
	if err := conn.SetLinger(0); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTCPConfigFailed, err)
	}

	// If keepAlive duration is not 0, set the keep alive options.
	if keepAlivePeriod != 0 {
		if err := conn.SetKeepAlive(true); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrTCPConfigFailed, err)
		}
		if err := conn.SetKeepAlivePeriod(keepAlivePeriod); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrTCPConfigFailed, err)
		}
	}

	// If handshakeTimeout is not 0, set the deadline for the secret connection handshake.
	if handshakeTimeout != 0 {
		conn.SetDeadline(time.Now().Add(handshakeTimeout))
	}

	// Secure the TCP connection and authenticate the remote signer.
	sconn, err := p2pconn.MakeSecretConnection(conn, localPrivKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSecretConnFailed, err)
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

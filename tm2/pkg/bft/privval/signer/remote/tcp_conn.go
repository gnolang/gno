package remote

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
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

	// SchemeAgnostic, when true, uses the scheme-agnostic SecretConnection
	// handshake (MakeSecretConnectionAny). Both ends of the privval channel
	// must agree on this setting. EXPLORATORY: the wire format of the auth
	// step diverges from the legacy ed25519-only format.
	SchemeAgnostic bool
}

// ConfigureTCPConnection configures the linger and keep alive options for a TCP connection.
// It also secures the connection and mutually authenticates with the remote peer using the
// provided localPrivKey and a list of authorizedKeys.
//
// When cfg.SchemeAgnostic is true, localPrivKey may be any crypto.PrivKey
// (ed25519 or secp256k1) and authorizedKeys is interpreted as []crypto.PubKey.
// When false (the legacy default), both are coerced to ed25519 types and the
// call fails if the coercion is impossible.
func ConfigureTCPConnection(
	conn *net.TCPConn,
	localPrivKey crypto.PrivKey,
	authorizedKeys []crypto.PubKey,
	cfg TCPConnConfig,
) (net.Conn, error) {
	if conn == nil {
		return nil, ErrNilConn
	}

	conn.SetLinger(0)

	if cfg.KeepAlivePeriod != 0 {
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(cfg.KeepAlivePeriod)
	}

	if cfg.HandshakeTimeout != 0 {
		conn.SetDeadline(time.Now().Add(cfg.HandshakeTimeout))
	}

	var (
		remotePubKey crypto.PubKey
		secured      net.Conn
	)

	if cfg.SchemeAgnostic {
		sconn, remPub, err := p2pconn.MakeSecretConnectionAny(conn, localPrivKey)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrSecretConnFailed, err)
		}
		secured = sconn
		remotePubKey = remPub
	} else {
		// Legacy ed25519-only path. Coerce inputs.
		localEd, ok := localPrivKey.(ed25519.PrivKeyEd25519)
		if !ok {
			return nil, fmt.Errorf("%w: legacy SecretConnection requires ed25519 private key, got %T",
				ErrSecretConnFailed, localPrivKey)
		}
		sconn, err := p2pconn.MakeSecretConnection(conn, localEd)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrSecretConnFailed, err)
		}
		secured = sconn
		remotePubKey = sconn.RemotePubKey()
	}

	if cfg.HandshakeTimeout != 0 {
		conn.SetDeadline(time.Time{})
	}

	if err := checkAuthorizedKeys(remotePubKey, authorizedKeys); err != nil {
		return nil, err
	}

	return secured, nil
}

// checkAuthorizedKeys checks if the public key of the remote peer is authorized.
func checkAuthorizedKeys(remotePubKey crypto.PubKey, authorizedKeys []crypto.PubKey) error {
	if len(authorizedKeys) == 0 {
		return nil
	}

	if slices.ContainsFunc(authorizedKeys, func(authorizedKey crypto.PubKey) bool {
		return authorizedKey.Equals(remotePubKey)
	}) {
		return nil
	}

	return ErrUnauthorizedPubKey
}

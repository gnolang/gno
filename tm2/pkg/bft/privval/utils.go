package privval

import (
	"fmt"
	"net"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

// IsConnTimeout returns a boolean indicating whether the error is known to
// report that a connection timeout occurred. This detects both fundamental
// network timeouts, as well as ErrConnTimeout errors.
func IsConnTimeout(err error) bool {
	switch errors.Cause(err).(type) {
	case EndpointTimeoutError:
		return true
	case timeoutError:
		return true
	default:
		return false
	}
}

// NewListener creates an UNIX or TCP listener using the corresponding listen address.
// TCP connection can be one-way or two-way authenticated using the listenerKey and authorizedKeys.
func NewListener(
	listenAddr string,
	listenerKey ed25519.PrivKeyEd25519,
	authorizedKeys []ed25519.PubKeyEd25519,
) (net.Listener, error) {
	var listener net.Listener

	protocol, address := osm.ProtocolAndAddress(listenAddr)
	ln, err := net.Listen(protocol, address)
	if err != nil {
		return nil, err
	}
	switch protocol {
	case "unix":
		listener = NewUnixListener(ln)
	case "tcp":
		listener = NewTCPListener(ln, listenerKey, authorizedKeys)
	default:
		return nil, fmt.Errorf(
			"wrong listen address: expected either 'tcp' or 'unix' protocols, got %s",
			protocol,
		)
	}

	return listener, nil
}

// NewDialer creates an UNIX or TCP dialer using the corresponding dialer address.
// TCP connection can be one-way or two-way authenticated using the dialerKey and authorizedKeys.
func NewDialer(
	dialAddr string,
	tcpTimeout time.Duration,
	dialerKey ed25519.PrivKeyEd25519,
	authorizedKeys []ed25519.PubKeyEd25519,
) (SocketDialer, error) {
	var dialer SocketDialer

	protocol, address := osm.ProtocolAndAddress(dialAddr)
	switch protocol {
	case "unix":
		dialer = DialUnixFn(address)
	case "tcp":
		dialer = DialTCPFn(address, tcpTimeout, dialerKey, authorizedKeys)
	default:
		return nil, fmt.Errorf(
			"wrong listen address: expected either 'tcp' or 'unix' protocols, got %s",
			protocol,
		)
	}

	return dialer, nil
}

// checkAuthorizedKeys checks if the public key of the remote peer is authorized.
func checkAuthorizedKeys(remotePubKey crypto.PubKey, authorizedKeys []ed25519.PubKeyEd25519) error {
	// If the whitelist is empty, skip the check
	if len(authorizedKeys) == 0 {
		return nil
	}

	// Check if the public key of the remote peer is authorized
	for _, key := range authorizedKeys {
		if remotePubKey.Equals(key) {
			return nil
		}
	}

	// If the public key was not found in the whitelist, return an error
	return errors.New("unauthorized public key")
}

package privval

import (
	"fmt"
	"log/slog"
	"net"
	"time"

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

// NewSignerListener creates a new SignerListenerEndpoint using the corresponding listen address.
func NewSignerListener(listenAddr string, logger *slog.Logger) (*SignerListenerEndpoint, error) {
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
		// TODO: persist this key so external signer can actually authenticate us
		listener = NewTCPListener(ln, ed25519.GenPrivKey())
	default:
		return nil, fmt.Errorf(
			"wrong listen address: expected either 'tcp' or 'unix' protocols, got %s",
			protocol,
		)
	}

	pvle := NewSignerListenerEndpoint(logger.With("module", "privval"), listener)

	return pvle, nil
}

// NewSignerDialer creates a new SignerDialerEndpoint using the corresponding dial address.
func NewSignerDialer(dialAddr string, tcpTimeout time.Duration, logger *slog.Logger) (*SignerDialerEndpoint, error) {
	var dialer SocketDialer

	protocol, address := osm.ProtocolAndAddress(dialAddr)
	switch protocol {
	case "unix":
		dialer = DialUnixFn(address)
	case "tcp":
		dialer = DialTCPFn(address, tcpTimeout, ed25519.GenPrivKey())
	default:
		return nil, fmt.Errorf(
			"wrong listen address: expected either 'tcp' or 'unix' protocols, got %s",
			protocol,
		)
	}

	pvde := NewSignerDialerEndpoint(logger.With("module", "privval"), dialer)

	return pvde, nil
}

package privval

import (
	"net"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	p2pconn "github.com/gnolang/gno/tm2/pkg/p2p/conn"
)

// Socket errors.
var (
	ErrDialRetryMax = errors.New("dialed maximum retries")
)

// SocketDialer dials a remote address and returns a net.Conn or an error.
type SocketDialer func() (net.Conn, error)

// DialTCPFn dials the given tcp addr, using the given timeoutReadWrite, dialerKey
// and authorizedKeys for the authenticated encrypted connection.
func DialTCPFn(
	addr string,
	timeoutReadWrite time.Duration,
	dialerKey ed25519.PrivKeyEd25519,
	authorizedKeys []ed25519.PubKeyEd25519,
) SocketDialer {
	return func() (net.Conn, error) {
		conn, err := osm.Connect(addr)
		if err != nil {
			return nil, err
		}

		deadline := time.Now().Add(timeoutReadWrite)
		if err = conn.SetDeadline(deadline); err != nil {
			return nil, err
		}

		secretConn, err := p2pconn.MakeSecretConnection(conn, dialerKey)
		if err != nil {
			return nil, err
		}

		// Check the public key of the remote peer against the authorized keys.
		if err := checkAuthorizedKeys(secretConn.RemotePubKey(), authorizedKeys); err != nil {
			secretConn.Close()
			return nil, err
		}

		return secretConn, nil
	}
}

// DialUnixFn dials the given unix socket.
func DialUnixFn(addr string) SocketDialer {
	return func() (net.Conn, error) {
		unixAddr := &net.UnixAddr{Name: addr, Net: "unix"}
		return net.DialUnix("unix", nil, unixAddr)
	}
}

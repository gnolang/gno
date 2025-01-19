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

// DialTCPFn dials the given tcp addr, using the given timeoutReadWrite and
// privKey for the authenticated encryption handshake.
func DialTCPFn(addr string, timeoutReadWrite time.Duration, privKey ed25519.PrivKeyEd25519) SocketDialer {
	return func() (net.Conn, error) {
		conn, err := osm.Connect(addr)
		if err != nil {
			return nil, err
		}

		deadline := time.Now().Add(timeoutReadWrite)
		if err = conn.SetDeadline(deadline); err != nil {
			return nil, err
		}

		return p2pconn.MakeSecretConnection(conn, privKey)
	}
}

// DialUnixFn dials the given unix socket.
func DialUnixFn(addr string) SocketDialer {
	return func() (net.Conn, error) {
		unixAddr := &net.UnixAddr{Name: addr, Net: "unix"}
		return net.DialUnix("unix", nil, unixAddr)
	}
}

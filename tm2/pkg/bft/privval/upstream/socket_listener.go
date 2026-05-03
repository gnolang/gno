package upstream

// socket_listener.go: TCP and Unix listeners with privval-specific
// timeouts and (for TCP) SecretConnection encryption + mutual auth via
// pubkey allowlist.
//
// Direct port of cometbft/privval/socket_listeners.go (CometBFT v0.39.1)
// with one tm2-specific adjustment: the allowlist check happens INSIDE
// Accept() rather than being delegated to the caller. This matches the
// existing gnokms pattern (tm2/pkg/bft/privval/signer/remote/tcp_conn.go::
// checkAuthorizedKeys) and prevents an unauthenticated peer's connection
// from reaching the SignerListenerEndpoint serve loop at all.
//
// UDS path intentionally bypasses SecretConnection — UDS already provides
// authenticated kernel-level isolation; encrypting it would be wasted.
// Same trade-off CometBFT makes.

import (
	"fmt"
	"net"
	"slices"
	"time"

	p2pconn "github.com/gnolang/gno/tm2/pkg/p2p/conn"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// ---- TCP Listener -----------------------------------------------------------

// TCPListenerOption sets an optional parameter on the TCPListener.
type TCPListenerOption func(*TCPListener)

// TCPListenerTimeoutAccept sets the per-Accept deadline. Zero disables.
func TCPListenerTimeoutAccept(timeout time.Duration) TCPListenerOption {
	return func(tl *TCPListener) { tl.timeoutAccept = timeout }
}

// TCPListenerTimeoutReadWrite sets the read/write deadline applied to
// each accepted connection.
func TCPListenerTimeoutReadWrite(timeout time.Duration) TCPListenerOption {
	return func(tl *TCPListener) { tl.timeoutReadWrite = timeout }
}

// TCPListener wraps a net.TCPListener with privval timeouts, SecretConnection
// encryption, and a mutual-auth allowlist. Accept returns an encrypted
// net.Conn whose remote pubkey is guaranteed to be in authorizedKeys.
type TCPListener struct {
	*net.TCPListener

	secretConnKey  ed25519.PrivKeyEd25519
	authorizedKeys []ed25519.PubKeyEd25519

	timeoutAccept    time.Duration
	timeoutReadWrite time.Duration
}

var _ net.Listener = (*TCPListener)(nil)

// NewTCPListener wraps ln. secretConnKey is the validator's identity
// (typically the node_id key). authorizedKeys is the allowlist of
// expected signer pubkeys (typically a single tmkms identity, or
// multiple cosigners for Horcrux). An empty allowlist accepts ANY peer
// that completes the SecretConnection handshake — equivalent to gnokms's
// "fail-open" mode and recommended only for dev/test.
func NewTCPListener(
	ln *net.TCPListener,
	secretConnKey ed25519.PrivKeyEd25519,
	authorizedKeys []ed25519.PubKeyEd25519,
	opts ...TCPListenerOption,
) *TCPListener {
	tl := &TCPListener{
		TCPListener:      ln,
		secretConnKey:    secretConnKey,
		authorizedKeys:   authorizedKeys,
		timeoutAccept:    time.Second * defaultTimeoutAcceptSeconds,
		timeoutReadWrite: time.Second * defaultTimeoutReadWriteSeconds,
	}
	for _, o := range opts {
		o(tl)
	}
	return tl
}

// Accept implements net.Listener. Sets the configured accept deadline,
// performs the SecretConnection handshake (which validates the remote
// peer signed a fresh challenge with its claimed pubkey), then verifies
// the remote pubkey is in the allowlist. Connections failing any of
// these are dropped before being returned.
func (ln *TCPListener) Accept() (net.Conn, error) {
	deadline := time.Now().Add(ln.timeoutAccept)
	if err := ln.SetDeadline(deadline); err != nil {
		return nil, err
	}

	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}

	// Apply read/write deadlines transparently for downstream readers.
	timeoutConn := newTimeoutConn(tc, ln.timeoutReadWrite)

	sconn, err := p2pconn.MakeSecretConnection(timeoutConn, ln.secretConnKey)
	if err != nil {
		_ = timeoutConn.Close()
		return nil, fmt.Errorf("upstream.TCPListener: SecretConnection handshake: %w", err)
	}

	if err := checkAuthorizedKey(sconn.RemotePubKey(), ln.authorizedKeys); err != nil {
		_ = sconn.Close()
		return nil, err
	}

	return sconn, nil
}

// ---- Unix Listener ----------------------------------------------------------

// UnixListenerOption sets an optional parameter on the UnixListener.
type UnixListenerOption func(*UnixListener)

// UnixListenerTimeoutAccept sets the per-Accept deadline.
func UnixListenerTimeoutAccept(timeout time.Duration) UnixListenerOption {
	return func(ul *UnixListener) { ul.timeoutAccept = timeout }
}

// UnixListenerTimeoutReadWrite sets the read/write deadline applied to
// each accepted connection.
func UnixListenerTimeoutReadWrite(timeout time.Duration) UnixListenerOption {
	return func(ul *UnixListener) { ul.timeoutReadWrite = timeout }
}

// UnixListener wraps a net.UnixListener with privval timeouts. UDS does
// NOT layer SecretConnection — kernel-level isolation suffices for
// same-host privval. Operators must protect the socket via filesystem
// perms.
type UnixListener struct {
	*net.UnixListener

	timeoutAccept    time.Duration
	timeoutReadWrite time.Duration
}

var _ net.Listener = (*UnixListener)(nil)

// NewUnixListener wraps ln (a *net.UnixListener).
func NewUnixListener(ln *net.UnixListener, opts ...UnixListenerOption) *UnixListener {
	ul := &UnixListener{
		UnixListener:     ln,
		timeoutAccept:    time.Second * defaultTimeoutAcceptSeconds,
		timeoutReadWrite: time.Second * defaultTimeoutReadWriteSeconds,
	}
	for _, o := range opts {
		o(ul)
	}
	return ul
}

// Accept implements net.Listener.
func (ln *UnixListener) Accept() (net.Conn, error) {
	deadline := time.Now().Add(ln.timeoutAccept)
	if err := ln.SetDeadline(deadline); err != nil {
		return nil, err
	}
	tc, err := ln.AcceptUnix()
	if err != nil {
		return nil, err
	}
	return newTimeoutConn(tc, ln.timeoutReadWrite), nil
}

// ---- Connection wrapper ----------------------------------------------------

// timeoutConn wraps a net.Conn to apply a read/write deadline on every op.
// Mirrors cometbft/privval/socket_listeners.go::timeoutConn.
type timeoutConn struct {
	net.Conn
	timeout time.Duration
}

func newTimeoutConn(conn net.Conn, timeout time.Duration) *timeoutConn {
	return &timeoutConn{Conn: conn, timeout: timeout}
}

// Read implements net.Conn.
func (c *timeoutConn) Read(b []byte) (int, error) {
	if c.timeout > 0 {
		if err := c.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
			return 0, err
		}
	}
	return c.Conn.Read(b)
}

// Write implements net.Conn.
func (c *timeoutConn) Write(b []byte) (int, error) {
	if c.timeout > 0 {
		if err := c.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
			return 0, err
		}
	}
	return c.Conn.Write(b)
}

// ---- allowlist helper ------------------------------------------------------

// checkAuthorizedKey returns an error if remotePubKey is not in
// authorizedKeys. Empty allowlist accepts all keys (caller's policy
// choice — useful for dev/test, dangerous for production; document
// loudly upstream).
func checkAuthorizedKey(remotePubKey ed25519.PubKeyEd25519, authorizedKeys []ed25519.PubKeyEd25519) error {
	if len(authorizedKeys) == 0 {
		return nil
	}
	if slices.ContainsFunc(authorizedKeys, func(k ed25519.PubKeyEd25519) bool {
		return k.Equals(remotePubKey)
	}) {
		return nil
	}
	return fmt.Errorf("upstream.TCPListener: remote pubkey %s not in allowlist", remotePubKey)
}

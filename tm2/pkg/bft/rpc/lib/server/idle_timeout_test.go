package rpcserver

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/log"
)

// startIdleTestServer starts the package's real HTTP server with the given
// config and returns the address to dial. The raw-TCP client used by these
// tests is deliberate: an http.Transport silently redials closed
// connections, which would mask the exact server-side connection lifecycle
// the tests assert.
func startIdleTestServer(t *testing.T, config *Config) string {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		fmt.Fprint(w, "ok")
	})

	l, err := Listen("tcp://127.0.0.1:0", config)
	require.NoError(t, err)
	t.Cleanup(func() { l.Close() })
	go StartHTTPServer(l, mux, log.NewTestingLogger(t), config)

	return l.Addr().String()
}

// rawGet writes one keep-alive HTTP/1.1 GET on conn and reads the complete
// response, leaving the connection open for reuse.
func rawGet(conn net.Conn, br *bufio.Reader) error {
	if _, err := fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: t\r\n\r\n"); err != nil {
		return err
	}
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		return err
	}
	_, err = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return err
}

// A keep-alive connection sitting idle between requests must survive past
// ReadTimeout when IdleTimeout says so. Without an IdleTimeout, net/http
// reuses ReadTimeout as the idle deadline, killing pooled connections held
// by reverse proxies (AWS ALB, nginx) and turning reuse into a 502 race.
func TestServerIdleConnOutlivesReadTimeout(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	config.ReadTimeout = 200 * time.Millisecond
	config.IdleTimeout = 2 * time.Second
	addr := startIdleTestServer(t, config)

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	require.NoError(t, err)
	defer conn.Close()
	br := bufio.NewReader(conn)

	require.NoError(t, rawGet(conn, br), "first request must succeed")

	// Idle well past ReadTimeout but well below IdleTimeout.
	time.Sleep(600 * time.Millisecond)

	require.NoError(t, rawGet(conn, br),
		"keep-alive reuse after an idle period > ReadTimeout must succeed")
}

// Idle connections must still be reaped once IdleTimeout elapses.
func TestServerIdleTimeoutReapsIdleConns(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	config.ReadTimeout = 200 * time.Millisecond
	config.IdleTimeout = 500 * time.Millisecond
	addr := startIdleTestServer(t, config)

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	require.NoError(t, err)
	defer conn.Close()
	br := bufio.NewReader(conn)

	require.NoError(t, rawGet(conn, br), "first request must succeed")

	// Block-read: the server must close the connection at IdleTimeout.
	start := time.Now()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = br.ReadByte()
	elapsed := time.Since(start)

	require.Error(t, err, "server must close the idle connection")
	require.GreaterOrEqual(t, elapsed, 300*time.Millisecond,
		"connection died before the IdleTimeout window")
	require.Less(t, elapsed, 4*time.Second,
		"connection outlived IdleTimeout by too much")
}

// A zero IdleTimeout preserves the historical behavior: net/http falls back
// to ReadTimeout as the idle deadline (mirrors http.Server#IdleTimeout).
func TestServerZeroIdleTimeoutFallsBackToReadTimeout(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	config.ReadTimeout = 300 * time.Millisecond
	config.IdleTimeout = 0
	addr := startIdleTestServer(t, config)

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	require.NoError(t, err)
	defer conn.Close()
	br := bufio.NewReader(conn)

	require.NoError(t, rawGet(conn, br), "first request must succeed")

	start := time.Now()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = br.ReadByte()
	elapsed := time.Since(start)

	require.Error(t, err, "idle connection must be closed at ReadTimeout")
	require.Less(t, elapsed, 2*time.Second,
		"zero IdleTimeout must fall back to ReadTimeout, not live longer")
}

// ReadTimeout must still bound in-flight request reads when IdleTimeout is
// set: a slow body is killed at ReadTimeout, not kept alive until
// IdleTimeout. This pins the DoS-hardening behavior the timeouts were
// introduced for (tendermint#2780).
func TestServerReadTimeoutStillBoundsSlowRequests(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	config.ReadTimeout = 300 * time.Millisecond
	config.IdleTimeout = 5 * time.Second
	addr := startIdleTestServer(t, config)

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	require.NoError(t, err)
	defer conn.Close()

	// Declare a body that never fully arrives.
	_, err = fmt.Fprintf(conn,
		"POST / HTTP/1.1\r\nHost: t\r\nContent-Length: 1000\r\n\r\npartial")
	require.NoError(t, err)

	start := time.Now()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = io.ReadAll(conn)
	elapsed := time.Since(start)

	require.NoError(t, err, "expected a server-side close, not a client-side deadline")
	require.Less(t, elapsed, 2500*time.Millisecond,
		"slow request must be killed at ReadTimeout, not IdleTimeout")
}

package server

import (
	"cmp"
	"fmt"
	"net"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	r "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote"
)

func (rss *RemoteSignerServer) setListener(listener net.Listener) error {
	rss.lock.Lock()
	defer rss.lock.Unlock()

	// If the listener is already set, close it.
	var err error
	if rss.listener != nil {
		err = rss.listener.Close()
	}

	rss.listener = listener

	return err
}

// clearConnection closes the served connection and wakes the accept loop. Given a conn, it acts only while
// conn is served, so a handler returning after Stop leaves open what a restarted server serves.
func (rss *RemoteSignerServer) clearConnection(conn net.Conn) error {
	rss.lock.Lock()
	defer rss.lock.Unlock()

	if rss.conn == nil || (conn != nil && rss.conn != conn) {
		return nil
	}

	err := rss.conn.Close()
	rss.conn = nil
	rss.idle.Broadcast()

	return err
}

// claimConnection makes conn the served connection, unless another one is served, the server stopped,
// or listener is not the server's, as for a connection a stopped server accepted.
func (rss *RemoteSignerServer) claimConnection(listener net.Listener, conn net.Conn) bool {
	rss.lock.Lock()
	defer rss.lock.Unlock()

	if rss.conn != nil || !rss.IsRunning() || rss.listener != listener {
		conn.Close()
		return false
	}

	rss.conn = conn

	return true
}

// waitIdle returns once no connection is served, which Stop also ensures.
func (rss *RemoteSignerServer) waitIdle() {
	rss.lock.Lock()
	defer rss.lock.Unlock()

	for rss.conn != nil {
		rss.idle.Wait()
	}
}

// serve accepts client connections and handles each in its own goroutine, so a stalled handshake holds up no other.
func (rss *RemoteSignerServer) serve(listener net.Listener) {
	rss.logger.Info("Start listening",
		"protocol", listener.Addr().Network(),
		"address", listener.Addr().String(),
	)

	// Peers opening connections faster than their handshakes expire can exhaust descriptors,
	// so a failed Accept waits before the next attempt rather than spin.
	var acceptDelay time.Duration

	for {
		// Accept incoming client connections.
		conn, err := listener.Accept()
		if err != nil {
			// If the server is still running, log the error and continue accepting connections.
			if rss.IsRunning() {
				rss.logger.Error("Failed to accept connection", "error", err)
				acceptDelay = min(max(2*acceptDelay, 5*time.Millisecond), time.Second)
				time.Sleep(acceptDelay)
				continue
			}
			break // Else, stop listening.
		}
		acceptDelay = 0
		rss.logger.Debug("Accepted new connection", "remote", conn.RemoteAddr())

		// One connection is served at a time: the next waits here, before its handshake, until the served
		// one closes, and later ones in the listener's backlog, so a TCP client's handshake times out.
		rss.waitIdle()

		go rss.handleConnection(listener, conn)
	}

	rss.logger.Info("Stop listening",
		"protocol", listener.Addr().Network(),
		"address", listener.Addr().String(),
	)
}

// handleConnection secures a TCP connection, then serves it unless another connection is served.
func (rss *RemoteSignerServer) handleConnection(listener net.Listener, conn net.Conn) {
	// If the connection is a TCP connection, configure and secure it.
	tcpConn, ok := conn.(*net.TCPConn)
	if ok {
		// The handshake keeps a deadline even with no response timeout, so a peer that never
		// completes it cannot hold its goroutine and descriptor for good.
		timeout := cmp.Or(rss.responseTimeout, DefaultResponseTimeout)

		tcpCfg := r.TCPConnConfig{
			KeepAlivePeriod:  rss.keepAlivePeriod,
			HandshakeTimeout: timeout * 2, // Double the response timeout for the handshake (send + receive).
		}

		// Configure and secure the TCP connection then authenticate the client.
		sconn, err := r.ConfigureTCPConnection(
			tcpConn,
			rss.serverPrivKey,
			rss.authorizedKeys,
			tcpCfg,
		)
		if err != nil {
			rss.logger.Error("Failed to configure TCP connection", "error", err)
			conn.Close() // Close the connection if its configuration failed.
			return
		}

		rss.logger.Debug("Configured TCP connection successfully")
		conn = sconn
	}

	// Handshakes that run while no connection is served race here: the first is served, the others closed.
	if !rss.claimConnection(listener, conn) {
		return
	}
	defer rss.clearConnection(conn)

	// Serve will run until the connection is closed or an error occurs while receiving
	// a request from or sending a response to the client.
	for {
		// Amino unmarshal target must be niled before unmarshaling.
		var request r.RemoteSignerMessage

		// Receive the request from the client and unmarshal it using amino.
		if _, err := amino.UnmarshalSizedReader(conn, &request, r.MaxMessageSize); err != nil {
			// Only log the error if the server is still running.
			if rss.IsRunning() {
				rss.logger.Error("Failed to receive request", "error", err)
			}
			break
		}

		// Handle the request and get the response.
		response := rss.handleRequest(request)

		// Set the deadline for the response sending.
		if rss.responseTimeout != 0 {
			conn.SetWriteDeadline(time.Now().Add(rss.responseTimeout))
		}

		// Marshal the response using amino then send it to the client.
		if _, err := amino.MarshalAnySizedWriter(conn, response); err != nil {
			// Only log the error if the server is still running.
			if rss.IsRunning() {
				rss.logger.Error("Failed to send response", "error", err)
			}
			break
		}

		rss.logger.Debug("Served request successfully", "request", fmt.Sprintf("%T", request))
	}
}

// handleRequest processes the incoming request and returns the response.
func (rss *RemoteSignerServer) handleRequest(request r.RemoteSignerMessage) r.RemoteSignerMessage {
	switch request := request.(type) {
	// PubKey request is proxied to the signer.
	case *r.PubKeyRequest:
		return &r.PubKeyResponse{PubKey: rss.signer.PubKey()}

		// Sign request is proxied to the signer.
	case *r.SignRequest:
		if signature, err := rss.signer.Sign(request.SignBytes); err != nil {
			return &r.SignResponse{Signature: nil, Error: &r.RemoteSignerError{Err: err.Error()}}
		} else {
			return &r.SignResponse{Signature: signature, Error: nil}
		}

	default:
		rss.logger.Error("Invalid request type", "type", fmt.Sprintf("%T", request))
		return nil
	}
}

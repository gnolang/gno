package server

import (
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	r "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote"
)

// closeListerners closes all listeners and remove them from the slice.
func (rss *RemoteSignerServer) closeListeners() error {
	var errors []string

	rss.listenersLock.Lock()
	defer rss.listenersLock.Unlock()

	// Iterate over all listeners and close them.
	for i := range rss.listeners {
		// Skip the listener if it's already closed.
		if rss.listeners[i] == nil {
			continue
		}

		// Close the listener and append the error to the slice if any.
		if err := rss.listeners[i].Close(); err != nil {
			errors = append(errors, fmt.Sprintf(
				"%s://%s: %v",
				rss.listeners[i].Addr().Network(),
				rss.listeners[i].Addr().String(),
				err,
			))
		}

		// Remove the listener from the slice since it's not usable anymore.
		rss.listeners[i] = nil
	}

	// Format the errors and return them if any.
	if len(errors) > 0 {
		return fmt.Errorf("closing listeners failed: [%s]", strings.Join(errors, ", "))
	}

	return nil
}

// addConnection adds a connection to the server.
func (rss *RemoteSignerServer) addConnection(conn net.Conn) {
	rss.connsLock.Lock()
	defer rss.connsLock.Unlock()
	rss.conns = append(rss.conns, conn)
}

// removeConnection removes a connection from the server.
func (rss *RemoteSignerServer) removeConnection(conn net.Conn) {
	rss.connsLock.Lock()
	defer rss.connsLock.Unlock()
	rss.conns = slices.DeleteFunc(rss.conns, func(entry net.Conn) bool { return entry == conn })
}

// closeConnections closes all connections.
func (rss *RemoteSignerServer) closeConnections() {
	rss.connsLock.RLock()
	defer rss.connsLock.RUnlock()

	// Iterate over all connections and close them.
	for _, conn := range rss.conns {
		conn.Close()
		// No need remove it from the slice since the serve goroutine will do it on exit.
	}
}

// listen starts the listener and accepts incoming connections.
func (rss *RemoteSignerServer) listen(listener net.Listener) {
	// The listener will run until the server is stopped.
	for {
		// Accept incoming client connections.
		conn, err := listener.Accept()
		if err != nil {
			// If the server is still running, log the error and continue accepting connections.
			if rss.IsRunning() {
				rss.logger.Error("Failed to accept connection", "error", err)
				continue
			}
			return
		}
		rss.logger.Debug("Accepted new connection", "remote", conn.RemoteAddr())

		// If the connection is a TCP connection, configure and secure it.
		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			// Configure and secure the TCP connection then authenticate the client.
			sconn, err := r.ConfigureTCPConnection(
				tcpConn,
				rss.serverPrivKey,
				rss.authorizedKeys,
				rss.keepAlivePeriod,
				rss.responseTimeout*2, // Double the response timeout for the handshake (send + receive).
			)
			if err != nil {
				rss.logger.Error("Failed to configure TCP connection", "error", err)
				conn.Close() // Close the connection if its configuration failed.
				continue
			}

			rss.logger.Debug("Configured TCP connection successfully")
			conn = sconn
		}

		// The connection is served in a separate goroutine which is added to the wait group.
		rss.wg.Add(1)
		go func(conn net.Conn) {
			defer rss.wg.Done()
			defer conn.Close() // Close the connection when the goroutine exits.

			rss.logger.Info("Connected to client",
				"protocol", conn.RemoteAddr().Network(),
				"address", conn.RemoteAddr().String(),
			)

			// Add the connection to the server then serve it.
			rss.addConnection(conn)
			rss.serve(conn)
			rss.removeConnection(conn)
		}(conn)
	}
}

// serve processes the incoming requests and sends the responses.
func (rss *RemoteSignerServer) serve(conn net.Conn) {
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
		response := rss.handle(request)

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

// handle processes the incoming request and returns the response.
func (rss *RemoteSignerServer) handle(request r.RemoteSignerMessage) r.RemoteSignerMessage {
	switch request := request.(type) {
	// PubKey request is proxied to the signer.
	case *r.PubKeyRequest:
		if pubKey, err := rss.signer.PubKey(); err != nil {
			return &r.PubKeyResponse{PubKey: nil, Error: &r.RemoteSignerError{Err: err.Error()}}
		} else {
			return &r.PubKeyResponse{PubKey: pubKey, Error: nil}
		}

		// Sign request is proxied to the signer.
	case *r.SignRequest:
		if signature, err := rss.signer.Sign(request.SignBytes); err != nil {
			return &r.SignResponse{Signature: nil, Error: &r.RemoteSignerError{Err: err.Error()}}
		} else {
			return &r.SignResponse{Signature: signature, Error: nil}
		}

		// Ping request is not related to the signer interface and is only used to confirm the connection.
	case *r.PingRequest:
		return &r.PingResponse{}

	default:
		rss.logger.Error("Invalid request type", "type", fmt.Sprintf("%T", request))
		return nil
	}
}

package server

import (
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

func (rss *RemoteSignerServer) setConnection(conn net.Conn) error {
	rss.lock.Lock()
	defer rss.lock.Unlock()

	// If the connection is already set, close it.
	var err error
	if rss.conn != nil {
		err = rss.conn.Close()
	}

	rss.conn = conn

	return err
}

// serve accepts incoming client connections and handles them.
func (rss *RemoteSignerServer) serve(listener net.Listener) {
	rss.logger.Info("Start listening",
		"protocol", listener.Addr().Network(),
		"address", listener.Addr().String(),
	)

	for {
		// Accept incoming client connections.
		conn, err := listener.Accept()
		if err != nil {
			// If the server is still running, log the error and continue accepting connections.
			if rss.IsRunning() {
				rss.logger.Error("Failed to accept connection", "error", err)
				continue
			}
			break // Else, stop listening.
		}
		rss.logger.Debug("Accepted new connection", "remote", conn.RemoteAddr())

		// If the connection is a TCP connection, configure and secure it.
		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			tcpCfg := r.TCPConnConfig{
				KeepAlivePeriod:  rss.keepAlivePeriod,
				HandshakeTimeout: rss.responseTimeout * 2, // Double the response timeout for the handshake (send + receive).
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
				continue
			}

			rss.logger.Debug("Configured TCP connection successfully")
			conn = sconn
		}

		// Start serving the connection.
		rss.handleConnection(conn)
	}

	rss.logger.Info("Stop listening",
		"protocol", listener.Addr().Network(),
		"address", listener.Addr().String(),
	)
}

// handleConnection handles the connection with the client.
func (rss *RemoteSignerServer) handleConnection(conn net.Conn) {
	rss.setConnection(conn)
	defer rss.setConnection(nil)

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

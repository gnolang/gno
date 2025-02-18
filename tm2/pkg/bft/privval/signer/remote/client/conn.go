package client

import (
	"net"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	r "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote"
)

// ensureConnection tries to establish a connection with the server.
func (rsc *RemoteSignerClient) ensureConnection() error {
	// If the connection is already established, return.
	if rsc.isConnected() {
		rsc.logger.Debug("already connected to server")
		return nil
	}

	// Try to establish a connection with the server, retrying if necessary.
	for try := 0; ; try++ {
		// Ensure the client is not closed.
		if rsc.isClosed() {
			return nil
		}

		// Dial the server.
		conn, err := net.DialTimeout(rsc.protocol, rsc.address, rsc.dialTimeout)
		if err != nil {
			rsc.logger.Warn("fail to dial",
				"protocol", rsc.protocol,
				"address", rsc.address,
				"error", err,
			)

			// If the maximum retries are not exceeded, log attempts count then retry.
			if rsc.dialMaxRetries > 0 && try < rsc.dialMaxRetries {
				rsc.logger.Info("retrying to connect", "try", try+1, "maxRetry", rsc.dialMaxRetries)
			} else if rsc.dialMaxRetries < 0 { // Retry indefinitely.
				rsc.logger.Info("retrying to connect", "try", try+1, "maxRetry", "unlimited")
			} else { // Max retries exceeded.
				return ErrMaxRetriesExceeded
			}

			// Wait for the retry interval before trying again.
			time.Sleep(rsc.dialRetryInterval)
			continue
		}
		rsc.logger.Debug("dial succeeded")

		// If the connection is a TCP connection, configure and secure it.
		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			// Configure and secure the TCP connection then authenticate the server.
			sconn, err := r.ConfigureTCPConnection(
				tcpConn,
				rsc.clientPrivKey,
				rsc.authorizedKeys,
				rsc.keepAlivePeriod,
				rsc.requestTimeout,
			)
			if err != nil {
				rsc.logger.Error("failed to configure TCP connection", "error", err)
				conn.Close() // Close the connection if its configuration failed.
				return err
			}

			rsc.logger.Debug("configured TCP connection successfully")
			conn = sconn
		}

		// Set the connection.
		rsc.setConnection(conn)
		rsc.logger.Info("connected to server", "protocol", rsc.protocol, "address", rsc.address)

		return nil
	}
}

// isConnected returns true if the client is connected to the server.
func (rsc *RemoteSignerClient) isConnected() bool {
	rsc.connLock.RLock()
	defer rsc.connLock.RUnlock()
	return rsc.conn != nil
}

// setConnection sets the connection to the server.
func (rsc *RemoteSignerClient) setConnection(conn net.Conn) error {
	rsc.connLock.Lock()
	defer rsc.connLock.Unlock()

	// Close the previous connection if it exists.
	var err error
	if rsc.conn != nil {
		err = rsc.conn.Close()
	}

	rsc.conn = conn

	return err
}

// send sends a request to the server and returns the response.
func (rsc *RemoteSignerClient) send(request r.RemoteSignerMessage) (r.RemoteSignerMessage, error) {
	var response r.RemoteSignerMessage

	// Ensure the client is not closed.
	if rsc.isClosed() {
		return nil, ErrClientAlreadyClosed
	}

	// This infinite loop ensures that if the connection is lost while sending the request
	// or receiving the response, the client will retry to establish the connection and
	// resend the request. This loop will break if the attempt to establish the connection
	// fails, if the client is closed or if the response is received successfully.
	for {
		// Ensure the connection is established.
		if err := rsc.ensureConnection(); err != nil {
			return nil, err
		}

		// Set the deadline for the request.
		if rsc.requestTimeout != 0 {
			rsc.conn.SetDeadline(time.Now().Add(rsc.requestTimeout))
		}

		// Marshal the request using amino then send it to the server.
		if _, err := amino.MarshalAnySizedWriter(rsc.conn, request); err != nil {
			rsc.logger.Warn("failed to send request", "error", err)
			rsc.setConnection(nil) // Close the connection if the sending the request failed.
			continue
		}

		// Receive the response from the server and unmarshal it using amino.
		if _, err := amino.UnmarshalSizedReader(rsc.conn, &response, r.MaxMessageSize); err != nil {
			rsc.logger.Warn("failed to receive response", "error", err)
			rsc.setConnection(nil) // Close the connection if the receiving the response failed.
			continue
		}

		return response, nil
	}
}

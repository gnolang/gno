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
		rsc.logger.Debug("Already connected to server")
		return nil
	}

	// Try to establish a connection with the server, retrying if necessary.
	for try := 0; ; try++ {
		// Ensure the client is not closed.
		if rsc.ctx.Err() != nil {
			return ErrClientAlreadyClosed
		}

		// Dial the server.
		conn, err := rsc.dialer.DialContext(rsc.ctx, rsc.protocol, rsc.address)
		if err != nil {
			rsc.logger.Warn("Failed to dial",
				"protocol", rsc.protocol,
				"address", rsc.address,
				"error", err,
			)

			// If the maximum retries are not exceeded, log attempts count then retry.
			if rsc.dialMaxRetries > 0 && try < rsc.dialMaxRetries {
				rsc.logger.Info("Retrying to connect", "try", try+1, "maxRetry", rsc.dialMaxRetries)
			} else if rsc.dialMaxRetries < 0 { // Retry indefinitely.
				rsc.logger.Info("Retrying to connect", "try", try+1, "maxRetry", "unlimited")
			} else { // Max retries exceeded.
				return ErrMaxRetriesExceeded
			}

			// Continue after the interval (retry) or if the dial context is done (exit).
			select {
			case <-time.After(rsc.dialRetryInterval):
			case <-rsc.ctx.Done():
			}
			continue
		}
		rsc.logger.Debug("Dial succeeded")

		// If the connection is a TCP connection, configure and secure it.
		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			tcpCfg := r.TCPConnConfig{
				KeepAlivePeriod:  rsc.keepAlivePeriod,
				HandshakeTimeout: rsc.requestTimeout,
			}

			// Configure and secure the TCP connection then authenticate the server.
			sconn, err := r.ConfigureTCPConnection(
				tcpConn,
				rsc.clientPrivKey,
				rsc.authorizedKeys,
				tcpCfg,
			)
			if err != nil {
				rsc.logger.Error("Failed to configure TCP connection", "error", err)
				conn.Close() // Close the connection if its configuration failed.
				return err
			}

			rsc.logger.Debug("Configured TCP connection successfully")
			conn = sconn
		}

		// Set the connection.
		rsc.setConnection(conn)
		rsc.logger.Info("Connected to server", "protocol", rsc.protocol, "address", rsc.address)

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
	// This infinite loop ensures that if the connection is lost while sending a request
	// or receiving a response, the client will retry establishing the connection and
	// resending the request. The loop will terminate if the connection attempt fails,
	// the client is closed, or the response is successfully received.
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
			rsc.logger.Warn("Failed to send request", "error", err)
			rsc.setConnection(nil) // Close the connection if the sending the request failed.
			continue
		}

		// Amino unmarshal target must be niled before unmarshaling.
		var response r.RemoteSignerMessage

		// Receive the response from the server and unmarshal it using amino.
		if _, err := amino.UnmarshalSizedReader(rsc.conn, &response, r.MaxMessageSize); err != nil {
			rsc.logger.Warn("Failed to receive response", "error", err)
			rsc.setConnection(nil) // Close the connection if receiving the response failed.
			continue
		}

		return response, nil
	}
}

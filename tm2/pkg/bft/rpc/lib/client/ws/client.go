package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"sync"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gorilla/websocket"
)

var (
	ErrTimedOut                  = errors.New("context timed out")
	ErrRequestResponseIDMismatch = errors.New("ws request / response ID mismatch")
	ErrInvalidBatchResponse      = errors.New("invalid ws batch response size")
)

type responseCh chan<- types.RPCResponses

// Client is a WebSocket client implementation
type Client struct {
	ctx           context.Context
	cancelCauseFn context.CancelCauseFunc

	conn *websocket.Conn

	logger  *slog.Logger
	backlog chan any // Either a single RPC request, or a batch of RPC requests

	requestMap    map[string]responseCh
	requestMapMux sync.Mutex
}

// NewClient initializes and creates a new WS RPC client
func NewClient(rpcURL string, opts ...Option) (*Client, error) {
	// Dial the RPC URL
	conn, _, err := websocket.DefaultDialer.Dial(rpcURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to dial RPC, %w", err)
	}

	c := &Client{
		conn:       conn,
		requestMap: make(map[string]responseCh),
		backlog:    make(chan any, 1),
		logger:     log.NewNoopLogger(),
	}

	ctx, cancelFn := context.WithCancelCause(context.Background())
	c.ctx = ctx
	c.cancelCauseFn = cancelFn

	// Apply the options
	for _, opt := range opts {
		opt(c)
	}

	go c.runReadRoutine(ctx)
	go c.runWriteRoutine(ctx)

	return c, nil
}

// SendRequest sends a single RPC request to the server
func (c *Client) SendRequest(ctx context.Context, request types.RPCRequest) (*types.RPCResponse, error) {
	// Create the response channel for the pipeline
	responseCh := make(chan types.RPCResponses, 1)

	// Generate a unique request ID hash
	requestHash := generateIDHash(request.ID.String())

	c.requestMapMux.Lock()
	c.requestMap[requestHash] = responseCh
	c.requestMapMux.Unlock()

	// Pipe the request to the backlog
	select {
	case <-ctx.Done():
		return nil, ErrTimedOut
	case <-c.ctx.Done():
		return nil, context.Cause(c.ctx)
	case c.backlog <- request:
	}

	// Wait for the response
	select {
	case <-ctx.Done():
		return nil, ErrTimedOut
	case <-c.ctx.Done():
		return nil, context.Cause(c.ctx)
	case response := <-responseCh:
		// Make sure the ID matches
		if response[0].ID != request.ID {
			// If response has an empty ID and an error, return the error instead of ID mismatch
			if (response[0].ID == nil || response[0].ID.String() == "") && response[0].Error != nil {
				return nil, response[0].Error
			}
			return nil, ErrRequestResponseIDMismatch
		}

		return &response[0], nil
	}
}

// SendBatch sends a batch of RPC requests to the server
func (c *Client) SendBatch(ctx context.Context, requests types.RPCRequests) (types.RPCResponses, error) {
	// Create the response channel for the pipeline
	responseCh := make(chan types.RPCResponses, 1)

	// Generate a unique request ID hash
	requestIDs := make([]string, 0, len(requests))

	for _, request := range requests {
		requestIDs = append(requestIDs, request.ID.String())
	}

	requestHash := generateIDHash(requestIDs...)

	c.requestMapMux.Lock()
	c.requestMap[requestHash] = responseCh
	c.requestMapMux.Unlock()

	// Pipe the request to the backlog
	select {
	case <-ctx.Done():
		return nil, ErrTimedOut
	case <-c.ctx.Done():
		return nil, context.Cause(c.ctx)
	case c.backlog <- requests:
	}

	// Wait for the response
	select {
	case <-ctx.Done():
		return nil, ErrTimedOut
	case <-c.ctx.Done():
		return nil, context.Cause(c.ctx)
	case responses := <-responseCh:
		// Make sure the length matches
		if len(responses) != len(requests) {
			return nil, ErrInvalidBatchResponse
		}

		// Make sure the IDs match
		for index, response := range responses {
			if requests[index].ID != response.ID {
				// If response has an empty ID and an error, return the error instead of ID mismatch
				if (response.ID == nil || response.ID.String() == "") && response.Error != nil {
					return nil, response.Error
				}
				return nil, ErrRequestResponseIDMismatch
			}
		}

		return responses, nil
	}
}

// generateIDHash generates a unique hash from the given IDs
func generateIDHash(ids ...string) string {
	hash := fnv.New128()

	for _, id := range ids {
		hash.Write([]byte(id))
	}

	return string(hash.Sum(nil))
}

// runWriteRoutine runs the client -> server write routine
func (c *Client) runWriteRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Debug("write context finished")

			return
		case item := <-c.backlog:
			// Write the JSON request to the server
			if err := c.conn.WriteJSON(item); err != nil {
				c.logger.Error("unable to send request", "err", err)

				continue
			}

			c.logger.Debug("successfully sent request", "request", item)
		}
	}
}

// runReadRoutine runs the client <- server read routine
func (c *Client) runReadRoutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.logger.Debug("read context finished")

			return
		default:
		}

		// Read the message from the active connection
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				c.logger.Error("failed to read response", "err", err)

				// Server dropped the connection, stop the client
				if err = c.closeWithCause(
					fmt.Errorf("server closed connection, %w", err),
				); err != nil {
					c.logger.Error("unable to gracefully close client", "err", err)
				}

				return
			}

			continue
		}

		var (
			responses    types.RPCResponses
			responseHash string
		)

		// Try to unmarshal as a batch of responses first
		if err := json.Unmarshal(data, &responses); err != nil {
			// Try to unmarshal as a single response
			var response types.RPCResponse

			if err := json.Unmarshal(data, &response); err != nil {
				c.logger.Error("failed to parse response", "err", err, "data", string(data))

				continue
			}

			// This is a single response, generate the unique ID
			responseHash = generateIDHash(response.ID.String())
			responses = types.RPCResponses{response}
		} else {
			// This is a batch response, generate the unique ID
			// from the combined IDs
			ids := make([]string, 0, len(responses))

			for _, response := range responses {
				ids = append(ids, response.ID.String())
			}

			responseHash = generateIDHash(ids...)
		}

		// Grab the response channel
		c.requestMapMux.Lock()
		ch := c.requestMap[responseHash]
		if ch == nil {
			c.requestMapMux.Unlock()
			c.logger.Error("response listener not set", "hash", responseHash, "responses", responses)

			continue
		}

		// Clear the entry for this ID
		delete(c.requestMap, responseHash)
		c.requestMapMux.Unlock()

		c.logger.Debug("received response", "hash", responseHash)

		// Alert the listener of the response
		select {
		case ch <- responses:
		default:
			c.logger.Warn("response listener timed out", "hash", responseHash)
		}
	}
}

// Close closes the WS client
func (c *Client) Close() error {
	return c.closeWithCause(nil)
}

// closeWithCause closes the client (and any open connection)
// with the given cause
func (c *Client) closeWithCause(err error) error {
	c.cancelCauseFn(err)

	return c.conn.Close()
}

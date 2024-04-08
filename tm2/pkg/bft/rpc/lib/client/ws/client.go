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
	"github.com/gorilla/websocket"
)

var errTimedOut = errors.New("context timed out")

type responseCh chan<- types.RPCResponses

// Client is a WebSocket client implementation
type Client struct {
	ctx      context.Context
	cancelFn context.CancelFunc

	conn *websocket.Conn

	logger *slog.Logger
	rpcURL string // the remote RPC URL of the node

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
		rpcURL:     rpcURL,
		conn:       conn,
		requestMap: make(map[string]responseCh),
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancelFn = cancelFn

	// Apply the options
	for _, opt := range opts {
		opt(c)
	}

	go c.runReadRoutine(ctx)
	go c.runWriteRoutine(ctx)

	return c, nil
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
		return nil, errTimedOut
	case c.backlog <- requests:
	}

	// Wait for the response
	select {
	case <-ctx.Done():
		return nil, errTimedOut
	case responses := <-responseCh:
		return responses, nil
	}
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
		return nil, errTimedOut
	case c.backlog <- request:
	}

	// Wait for the response
	select {
	case <-ctx.Done():
		return nil, errTimedOut
	case response := <-responseCh:
		return &response[0], nil
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
			// Read the message from the active connection
			_, data, err := c.conn.ReadMessage() // TODO check message type
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					c.logger.Error("failed to read response", "err", err)

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
}

// Close closes the WS client
func (c *Client) Close() {
	c.cancelFn()
}

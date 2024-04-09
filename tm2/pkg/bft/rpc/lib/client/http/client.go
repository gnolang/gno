package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

// Client is an HTTP client implementation
type Client struct {
	rpcURL string // the remote RPC URL of the node

	client *http.Client
}

// NewClient initializes and creates a new HTTP RPC client
func NewClient(rpcURL string) (*Client, error) {
	// Parse the RPC URL
	address, err := toClientAddress(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("invalid RPC URL, %w", err)
	}

	c := &Client{
		rpcURL: address,
		client: defaultHTTPClient(rpcURL),
	}

	return c, nil
}

// SendRequest sends a single RPC request to the server
func (c *Client) SendRequest(ctx context.Context, request types.RPCRequest) (*types.RPCResponse, error) {
	return sendRequestCommon[types.RPCRequest, *types.RPCResponse](ctx, c.client, c.rpcURL, request)
}

// SendBatch sends a single RPC batch request to the server
func (c *Client) SendBatch(ctx context.Context, requests types.RPCRequests) (types.RPCResponses, error) {
	return sendRequestCommon[types.RPCRequests, types.RPCResponses](ctx, c.client, c.rpcURL, requests)
}

type (
	requestType interface {
		types.RPCRequest | types.RPCRequests
	}

	responseType interface {
		*types.RPCResponse | types.RPCResponses
	}
)

// sendRequestCommon executes the common request sending
func sendRequestCommon[T requestType, R responseType](
	ctx context.Context,
	client *http.Client,
	rpcURL string,
	request T,
) (R, error) {
	// Marshal the request
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("unable to JSON-marshal the request, %w", err)
	}

	// Craft the request
	req, err := http.NewRequest(
		http.MethodPost,
		rpcURL,
		bytes.NewBuffer(requestBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create request, %w", err)
	}

	// Set the header content type
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	httpResponse, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("unable to send request, %w", err)
	}
	defer httpResponse.Body.Close() //nolint: errcheck

	// Parse the response code
	if !isOKStatus(httpResponse.StatusCode) {
		return nil, fmt.Errorf("invalid status code received, %d", httpResponse.StatusCode)
	}

	// Parse the response body
	responseBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body, %w", err)
	}

	var response R

	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response body, %w", err)
	}

	return response, nil
}

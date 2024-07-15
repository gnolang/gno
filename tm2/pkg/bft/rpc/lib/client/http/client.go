package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

const (
	protoHTTP  = "http"
	protoHTTPS = "https"
	protoTCP   = "tcp"
)

var (
	ErrRequestResponseIDMismatch = errors.New("http request / response ID mismatch")
	ErrInvalidBatchResponse      = errors.New("invalid http batch response size")
)

type AuthInfo struct {
	Username string
	Password string
}

// Client is an HTTP client implementation
type Client struct {
	rpcURL string // the remote RPC URL of the node

	authInfo *AuthInfo
	client   *http.Client
}

// NewClient initializes and creates a new HTTP RPC client
func NewClient(rpcURL string) (*Client, error) {
	// Parse the RPC URL
	address, err := toClientAddress(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("invalid RPC URL, %w", err)
	}

	parseAuthInfo(rpcURL)

	c := &Client{
		rpcURL:   address,
		authInfo: &AuthInfo{},
		client:   defaultHTTPClient(rpcURL),
	}

	return c, nil
}

// SendRequest sends a single RPC request to the server
func (c *Client) SendRequest(ctx context.Context, request types.RPCRequest) (*types.RPCResponse, error) {
	// Send the request
	response, err := sendRequestCommon[types.RPCRequest, *types.RPCResponse](ctx, c.client, c.rpcURL, c.authInfo, request)
	if err != nil {
		return nil, err
	}

	// Make sure the ID matches
	if response.ID != response.ID {
		return nil, ErrRequestResponseIDMismatch
	}

	return response, nil
}

// SendBatch sends a single RPC batch request to the server
func (c *Client) SendBatch(ctx context.Context, requests types.RPCRequests) (types.RPCResponses, error) {
	// Send the batch
	responses, err := sendRequestCommon[types.RPCRequests, types.RPCResponses](ctx, c.client, c.rpcURL, c.authInfo, requests)
	if err != nil {
		return nil, err
	}

	// Make sure the length matches
	if len(responses) != len(requests) {
		return nil, ErrInvalidBatchResponse
	}

	// Make sure the IDs match
	for index, response := range responses {
		if requests[index].ID != response.ID {
			return nil, ErrRequestResponseIDMismatch
		}
	}

	return responses, nil
}

// Close has no effect on an HTTP client
func (c *Client) Close() error {
	return nil
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
	authInfo *AuthInfo,
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

	// Set the basic authentication
	if authInfo != nil {
		req.SetBasicAuth(authInfo.Username, authInfo.Password)
	}

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

// DefaultHTTPClient is used to create an http client with some default parameters.
// We overwrite the http.Client.Dial so we can do http over tcp or unix.
// remoteAddr should be fully featured (eg. with tcp:// or unix://)
func defaultHTTPClient(remoteAddr string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			// Set to true to prevent GZIP-bomb DoS attacks
			DisableCompression: true,
			DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
				return makeHTTPDialer(remoteAddr)(network, addr)
			},
		},
	}
}

func makeHTTPDialer(remoteAddr string) func(string, string) (net.Conn, error) {
	protocol, address := parseRemoteAddr(remoteAddr)

	// net.Dial doesn't understand http/https, so change it to TCP
	switch protocol {
	case protoHTTP, protoHTTPS:
		protocol = protoTCP
	}

	return func(proto, addr string) (net.Conn, error) {
		return net.Dial(protocol, address)
	}
}

// protocol - client's protocol (for example, "http", "https", "wss", "ws", "tcp")
// trimmedS - rest of the address (for example, "192.0.2.1:25", "[2001:db8::1]:80") with "/" replaced with "."
func toClientAddrAndParse(remoteAddr string) (string, string) {
	protocol, address := parseRemoteAddr(remoteAddr)

	// protocol to use for http operations, to support both http and https
	var clientProtocol string
	// default to http for unknown protocols (ex. tcp)
	switch protocol {
	case protoHTTP, protoHTTPS:
		clientProtocol = protocol
	default:
		clientProtocol = protoHTTP
	}

	// replace / with . for http requests (kvstore domain)
	trimmedAddress := strings.Replace(address, "/", ".", -1)

	return clientProtocol, trimmedAddress
}

func toClientAddress(remoteAddr string) (string, error) {
	clientProtocol, trimmedAddress := toClientAddrAndParse(remoteAddr)

	return clientProtocol + "://" + trimmedAddress, nil
}

// network - name of the network (for example, "tcp", "unix")
// s - rest of the address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
// TODO: Deprecate support for IP:PORT or /path/to/socket
func parseRemoteAddr(remoteAddr string) (string, string) {
	parts := strings.SplitN(remoteAddr, "://", 2)

	var protocol, address string
	switch len(parts) {
	case 1:
		// default to tcp if nothing specified
		protocol, address = protoTCP, remoteAddr
	case 2:
		protocol, address = parts[0], parts[1]
	}

	// Append default ports if not specified
	if !strings.Contains(address, ":") {
		switch protocol {
		case protoHTTPS:
			address += ":443"
		case protoHTTP, protoTCP:
			address += ":80"
		default: // noop
		}
	}

	return protocol, address
}

func parseAuthInfo(remoteAddr string) (AuthInfo, error) {
	parsedURL, err := url.Parse(remoteAddr)
	if err != nil {
		return AuthInfo{}, err
	}

	username := parsedURL.User.Username()
	password, _ := parsedURL.User.Password()

	return AuthInfo{
		Username: username,
		Password: password,
	}, nil
}

// isOKStatus returns a boolean indicating if the response
// status code is between 200 and 299 (inclusive)
func isOKStatus(code int) bool { return code >= 200 && code <= 299 }

package rpcclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/random"
)

const (
	protoHTTP  = "http"
	protoHTTPS = "https"
	protoWSS   = "wss"
	protoWS    = "ws"
	protoTCP   = "tcp"
)

// HTTPClient is a common interface for JSONRPCClient and URIClient.
type HTTPClient interface {
	Call(method string, params map[string]any, result any) error
}

// protocol - client's protocol (for example, "http", "https", "wss", "ws", "tcp")
// trimmedS - rest of the address (for example, "192.0.2.1:25", "[2001:db8::1]:80") with "/" replaced with "."
func toClientAddrAndParse(remoteAddr string) (network string, trimmedS string, err error) {
	protocol, address, err := parseRemoteAddr(remoteAddr)
	if err != nil {
		return "", "", err
	}

	// protocol to use for http operations, to support both http and https
	var clientProtocol string
	// default to http for unknown protocols (ex. tcp)
	switch protocol {
	case protoHTTP, protoHTTPS, protoWS, protoWSS:
		clientProtocol = protocol
	default:
		clientProtocol = protoHTTP
	}

	// replace / with . for http requests (kvstore domain)
	trimmedAddress := strings.Replace(address, "/", ".", -1)
	return clientProtocol, trimmedAddress, nil
}

func toClientAddress(remoteAddr string) (string, error) {
	clientProtocol, trimmedAddress, err := toClientAddrAndParse(remoteAddr)
	if err != nil {
		return "", err
	}
	return clientProtocol + "://" + trimmedAddress, nil
}

// network - name of the network (for example, "tcp", "unix")
// s - rest of the address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
// TODO: Deprecate support for IP:PORT or /path/to/socket
func parseRemoteAddr(remoteAddr string) (network string, s string, err error) {
	parts := strings.SplitN(remoteAddr, "://", 2)
	var protocol, address string
	switch {
	case len(parts) == 1:
		// default to tcp if nothing specified
		protocol, address = protoTCP, remoteAddr
	case len(parts) == 2:
		protocol, address = parts[0], parts[1]
	default:
		return "", "", fmt.Errorf("invalid addr: %s", remoteAddr)
	}

	return protocol, address, nil
}

func makeErrorDialer(err error) func(string, string) (net.Conn, error) {
	return func(_ string, _ string) (net.Conn, error) {
		return nil, err
	}
}

func makeHTTPDialer(remoteAddr string) func(string, string) (net.Conn, error) {
	protocol, address, err := parseRemoteAddr(remoteAddr)
	if err != nil {
		return makeErrorDialer(err)
	}

	// net.Dial doesn't understand http/https, so change it to TCP
	switch protocol {
	case protoHTTP, protoHTTPS:
		protocol = protoTCP
	}

	return func(proto, addr string) (net.Conn, error) {
		return net.Dial(protocol, address)
	}
}

// DefaultHTTPClient is used to create an http client with some default parameters.
// We overwrite the http.Client.Dial so we can do http over tcp or unix.
// remoteAddr should be fully featured (eg. with tcp:// or unix://)
func DefaultHTTPClient(remoteAddr string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			// Set to true to prevent GZIP-bomb DoS attacks
			DisableCompression: true,
			Dial:               makeHTTPDialer(remoteAddr),
		},
	}
}

// ------------------------------------------------------------------------------------

// JSONRPCClient takes params as a slice
type JSONRPCClient struct {
	address  string
	client   *http.Client
	idPrefix types.JSONRPCStringID
}

// RPCCaller implementers can facilitate calling the JSON RPC endpoint.
type RPCCaller interface {
	Call(method string, params map[string]any, result any) error
}

// WrappedRPCRequest encapsulates a single buffered request, as well as its
// anticipated response structure
type WrappedRPCRequest struct {
	request types.RPCRequest
	result  any // The result will be deserialized into this object (Amino)
}

type WrappedRPCRequests []*WrappedRPCRequest

func (w *WrappedRPCRequest) extractRPCRequest() types.RPCRequest {
	return w.request
}

func (w *WrappedRPCRequests) extractRPCRequests() types.RPCRequests {
	requests := make([]types.RPCRequest, 0, len(*w))

	for _, wrappedRequest := range *w {
		requests = append(requests, wrappedRequest.request)
	}

	return requests
}

var (
	_ RPCCaller   = (*JSONRPCClient)(nil)
	_ BatchClient = (*JSONRPCClient)(nil)
)

// NewJSONRPCClient returns a JSONRPCClient pointed at the given address.
func NewJSONRPCClient(remote string) *JSONRPCClient {
	return NewJSONRPCClientWithHTTPClient(remote, DefaultHTTPClient(remote))
}

// NewJSONRPCClientWithHTTPClient returns a JSONRPCClient pointed at the given address using a custom http client
// The function panics if the provided client is nil or remote is invalid.
func NewJSONRPCClientWithHTTPClient(remote string, client *http.Client) *JSONRPCClient {
	if client == nil {
		panic("nil http.Client provided")
	}

	clientAddress, err := toClientAddress(remote)
	if err != nil {
		panic(fmt.Sprintf("invalid remote %s: %s", remote, err))
	}

	return &JSONRPCClient{
		address:  clientAddress,
		client:   client,
		idPrefix: types.JSONRPCStringID("jsonrpc-client-" + random.RandStr(8)),
	}
}

// Call will send the request for the given method through to the RPC endpoint
// immediately, without buffering of requests.
func (c *JSONRPCClient) Call(method string, params map[string]any, result any) error {
	id := generateRequestID(c.idPrefix)

	request, err := types.MapToRequest(id, method, params)
	if err != nil {
		return err
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}
	requestBuf := bytes.NewBuffer(requestBytes)
	httpResponse, err := c.client.Post(c.address, "text/json", requestBuf)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close() //nolint: errcheck

	if !statusOK(httpResponse.StatusCode) {
		return errors.New("server at '%s' returned %s", c.address, httpResponse.Status)
	}

	responseBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}

	var response types.RPCResponse

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return errors.Wrap(err, "error unmarshalling rpc response")
	}

	if response.Error != nil {
		return errors.Wrap(response.Error, "response error")
	}

	return unmarshalResponseIntoResult(&response, id, result)
}

func (c *JSONRPCClient) SendBatch(_ context.Context, wrappedRequests WrappedRPCRequests) (types.RPCResponses, error) {
	requests := make(types.RPCRequests, 0, len(wrappedRequests))
	for _, request := range wrappedRequests {
		requests = append(requests, request.request)
	}

	// serialize the array of requests into a single JSON object
	requestBytes, err := json.Marshal(requests)
	if err != nil {
		return nil, err
	}

	httpResponse, err := c.client.Post(c.address, "text/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close() //nolint: errcheck

	if !statusOK(httpResponse.StatusCode) {
		return nil, errors.New("server at '%s' returned %s", c.address, httpResponse.Status)
	}

	responseBytes, err := io.ReadAll(httpResponse.Body)

	var responses types.RPCResponses

	if err = json.Unmarshal(responseBytes, &responses); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling rpc responses")
	}

	return responses, nil
}

func (c *JSONRPCClient) GetIDPrefix() types.JSONRPCID {
	return c.idPrefix
}

func unmarshalResponseIntoResult(response *types.RPCResponse, expectedID types.JSONRPCID, result any) error {
	// Read response.  If rpc/core/types is imported, the result will unmarshal
	// into the correct type.
	// From the JSON-RPC 2.0 spec:
	//  idPrefix: It MUST be the same as the value of the idPrefix member in the Request Object.
	if err := validateResponseID(response, expectedID); err != nil {
		return err
	}

	// Unmarshal the RawMessage into the result.
	if err := amino.UnmarshalJSON(response.Result, result); err != nil {
		return errors.Wrap(err, "error unmarshalling rpc response result")
	}

	return nil
}

func unmarshalResponsesIntoResults(requests types.RPCRequests, responses types.RPCResponses, results []any) error {
	// No response error checking here as there may be a mixture of successful
	// and unsuccessful responses
	if len(results) != len(responses) {
		return fmt.Errorf("expected %d result objects into which to inject responses, but got %d", len(responses), len(results))
	}

	for i, response := range responses {
		response := response
		// From the JSON-RPC 2.0 spec:
		//  idPrefix: It MUST be the same as the value of the idPrefix member in the Request Object.

		// This validation is super sketchy. Why do this here?
		// This validation passes iff the server returns batch responses
		// in the same order as the batch request
		if err := validateResponseID(&response, requests[i].ID); err != nil {
			return errors.Wrap(err, "failed to validate response ID in response %d", i)
		}
		if err := amino.UnmarshalJSON(responses[i].Result, results[i]); err != nil {
			return errors.Wrap(err, "error unmarshalling rpc response result")
		}
	}

	return nil
}

func validateResponseID(res *types.RPCResponse, expectedID types.JSONRPCID) error {
	_, isNumValue := expectedID.(types.JSONRPCIntID)
	stringValue, isStringValue := expectedID.(types.JSONRPCStringID)

	if !isNumValue && !isStringValue {
		return errors.New("invalid expected ID")
	}

	// we only validate a response ID if the expected ID is non-empty
	if isStringValue && len(stringValue) == 0 {
		return nil
	}

	if res.ID == nil {
		return errors.New("missing ID in response")
	}

	if expectedID != res.ID {
		return fmt.Errorf("response ID (%s) does not match request ID (%s)", res.ID, expectedID)
	}

	return nil
}

func argsToURLValues(args map[string]any) (url.Values, error) {
	values := make(url.Values)
	if len(args) == 0 {
		return values, nil
	}
	err := argsToJSON(args)
	if err != nil {
		return nil, err
	}
	for key, val := range args {
		values.Set(key, val.(string))
	}
	return values, nil
}

func argsToJSON(args map[string]any) error {
	for k, v := range args {
		rt := reflect.TypeOf(v)
		isByteSlice := rt.Kind() == reflect.Slice && rt.Elem().Kind() == reflect.Uint8
		if isByteSlice {
			bytes := reflect.ValueOf(v).Bytes()
			args[k] = fmt.Sprintf("0x%X", bytes)
			continue
		}

		data, err := amino.MarshalJSON(v)
		if err != nil {
			return err
		}
		args[k] = string(data)
	}
	return nil
}

func statusOK(code int) bool { return code >= 200 && code <= 299 }

// generateRequestID generates a unique request ID, using the prefix
// Assuming this is sufficiently random, there shouldn't be any problems.
// However, using uuid for any kind of ID generation is always preferred
func generateRequestID(prefix types.JSONRPCID) types.JSONRPCID {
	return types.JSONRPCStringID(fmt.Sprintf("%s-%s", prefix, random.RandStr(8)))
}

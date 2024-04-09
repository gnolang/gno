package rpcclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	http2 "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/http"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

// URI takes params as a map
type URIClient struct {
	address string
	client  *http.Client
}

// The function panics if the provided remote is invalid.
func NewURIClient(remote string) *URIClient {
	clientAddress, err := http2.toClientAddress(remote)
	if err != nil {
		panic(fmt.Sprintf("invalid remote %s: %s", remote, err))
	}
	return &URIClient{
		address: clientAddress,
		client:  http2.DefaultHTTPClient(remote),
	}
}

func (c *URIClient) Call(method string, params map[string]any, result any) error {
	values, err := http2.argsToURLValues(params)
	if err != nil {
		return err
	}
	// log.Info(Fmt("URI request to %v (%v): %v", c.address, method, values))
	resp, err := c.client.PostForm(c.address+"/"+method, values)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint: errcheck

	if !http2.statusOK(resp.StatusCode) {
		return errors.New("server at '%s' returned %s", c.address, resp.Status)
	}

	responseBytes, err := io.ReadAll(resp.Body)
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

	return http2.unmarshalResponseIntoResult(&response, types.JSONRPCStringID(""), result)
}

package gnoclient

// Client represents the Gno.land RPC API client.
type Client struct {
	Remote  string
	ChainID string
}

// TODO: port existing code, i.e. faucet?
// TODO: create right now a tm2 generic go client and a gnovm generic go client?
// TODO: Command: Call
// TODO: Command: Send
// TODO: Command: AddPkg
// TODO: Command: Query
// TODO: Command: Eval
// TODO: Command: Exec
// TODO: Command: Package
// TODO: Command: QFile
// TODO: examples and unit tests
// TODO: Mock
// TODO: alternative configuration (pass existing websocket?)
// TODO: minimal go.mod to make it light to import

func (c *Client) ApplyDefaults() {
	if c.Remote == "" {
		c.Remote = "127.0.0.1:26657"
	}
	if c.ChainID == "" {
		c.ChainID = "devnet"
	}
}

// Request performs an API request and returns the response body.
func (c *Client) Request(method, endpoint string, params map[string]interface{}) ([]byte, error) {
	panic("not implemented")
}

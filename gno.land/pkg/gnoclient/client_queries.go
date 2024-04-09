package gnoclient

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// QueryCfg contains configuration options for performing ABCI queries.
type QueryCfg struct {
	Path                       string // Query path
	Data                       []byte // Query data
	rpcclient.ABCIQueryOptions        // ABCI query options
}

// Query performs a generic query on the blockchain.
func (c *Client) Query(cfg QueryCfg) (*ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}
	qres, err := c.RPCClient.ABCIQueryWithOptions(cfg.Path, cfg.Data, cfg.ABCIQueryOptions)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}

	if qres.Response.Error != nil {
		return qres, errors.Wrap(qres.Response.Error, "deliver transaction failed: log:%s", qres.Response.Log)
	}

	return qres, nil
}

// QueryAccount retrieves account information for a given address.
func (c *Client) QueryAccount(addr crypto.Address) (*std.BaseAccount, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("auth/accounts/%s", crypto.AddressToBech32(addr))
	data := []byte{}

	qres, err := c.RPCClient.ABCIQuery(path, data)
	if err != nil {
		return nil, nil, errors.Wrap(err, "query account")
	}
	if qres.Response.Data == nil || len(qres.Response.Data) == 0 || string(qres.Response.Data) == "null" {
		return nil, nil, std.ErrUnknownAddress("unknown address: " + crypto.AddressToBech32(addr))
	}

	var qret struct{ BaseAccount std.BaseAccount }
	err = amino.UnmarshalJSON(qres.Response.Data, &qret)
	if err != nil {
		return nil, nil, err
	}

	return &qret.BaseAccount, qres, nil
}

// QueryAppVersion retrieves information about the app version
func (c *Client) QueryAppVersion() (string, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return "", nil, err
	}

	path := ".app/version"
	data := []byte{}

	qres, err := c.RPCClient.ABCIQuery(path, data)
	if err != nil {
		return "", nil, errors.Wrap(err, "query app version")
	}

	version := string(qres.Response.Value)
	return version, qres, nil
}

// Render calls the Render function for pkgPath with optional args. The pkgPath should
// include the prefix like "gno.land/". This is similar to using a browser URL
// <testnet>/<pkgPath>:<args> where <pkgPath> doesn't have the prefix like "gno.land/".
func (c *Client) Render(pkgPath string, args string) (string, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return "", nil, err
	}

	path := "vm/qrender"
	data := []byte(fmt.Sprintf("%s\n%s", pkgPath, args))

	qres, err := c.RPCClient.ABCIQuery(path, data)
	if err != nil {
		return "", nil, errors.Wrap(err, "query render")
	}
	if qres.Response.Error != nil {
		return "", nil, errors.Wrap(qres.Response.Error, "Render failed: log:%s", qres.Response.Log)
	}

	return string(qres.Response.Data), qres, nil
}

// QEval evaluates the given expression with the realm code at pkgPath. The pkgPath should
// include the prefix like "gno.land/". The expression is usually a function call like
// "GetBoardIDFromName(\"testboard\")". The return value is a typed expression like
// "(1 gno.land/r/demo/boards.BoardID)\n(true bool)".
func (c *Client) QEval(pkgPath string, expression string) (string, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return "", nil, err
	}

	path := "vm/qeval"
	data := []byte(fmt.Sprintf("%s\n%s", pkgPath, expression))

	qres, err := c.RPCClient.ABCIQuery(path, data)
	if err != nil {
		return "", nil, errors.Wrap(err, "query qeval")
	}
	if qres.Response.Error != nil {
		return "", nil, errors.Wrap(qres.Response.Error, "QEval failed: log:%s", qres.Response.Log)
	}

	return string(qres.Response.Data), qres, nil
}

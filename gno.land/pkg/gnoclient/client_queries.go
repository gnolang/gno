package gnoclient

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// QueryCfg contains configuration options for performing queries.
type QueryCfg struct {
	Path                       string // Query path
	Data                       []byte // Query data
	rpcclient.ABCIQueryOptions        // ABCI query options
}

// Query performs a generic query on the blockchain.
func (c Client) Query(cfg QueryCfg) (*ctypes.ResultABCIQuery, error) {
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
func (c Client) QueryAccount(addr string) (*std.BaseAccount, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("auth/accounts/%s", addr)
	data := []byte{}

	qres, err := c.RPCClient.ABCIQuery(path, data)
	if err != nil {
		return nil, nil, errors.Wrap(err, "query account")
	}

	var qret struct{ BaseAccount std.BaseAccount }
	err = amino.UnmarshalJSON(qres.Response.Data, &qret)
	if err != nil {
		return nil, nil, err
	}

	return &qret.BaseAccount, qres, nil
}

func (c Client) QueryAppVersion() (string, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return "", nil, err
	}

	path := ".app/version"
	data := []byte{}

	qres, err := c.RPCClient.ABCIQuery(path, data)
	if err != nil {
		return "", nil, errors.Wrap(err, "query account")
	}

	version := string(qres.Response.Value)
	return version, qres, nil
}

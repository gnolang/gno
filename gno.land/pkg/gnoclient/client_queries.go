package gnoclient

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var ErrInvalidBlockHeight = errors.New("invalid block height provided")

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
	qres, err := c.RPCClient.ABCIQueryWithOptions(context.Background(), cfg.Path, cfg.Data, cfg.ABCIQueryOptions)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}

	if qres.Response.Error != nil {
		return qres, errors.Wrapf(qres.Response.Error, "deliver transaction failed: log:%s", qres.Response.Log)
	}

	return qres, nil
}

// QueryAccount retrieves account information for a given address.
//
// The returned account's Coins field holds only the chain's gas denom. Every
// other denom lives in its own store key; query bank/balances for the full set.
func (c *Client) QueryAccount(addr crypto.Address) (*std.BaseAccount, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("auth/accounts/%s", crypto.AddressToBech32(addr))
	data := []byte{}

	qres, err := c.RPCClient.ABCIQuery(context.Background(), path, data)
	if err != nil {
		return nil, nil, errors.Wrap(err, "query account")
	}
	if len(qres.Response.Data) == 0 || string(qres.Response.Data) == "null" {
		return nil, nil, std.ErrUnknownAddress("unknown address: " + crypto.AddressToBech32(addr))
	}

	var qret gnoland.GnoAccount
	err = amino.UnmarshalJSON(qres.Response.Data, &qret)
	if err != nil {
		return nil, nil, err
	}

	return &qret.BaseAccount, qres, nil
}

// QueryBalance retrieves every coin balance held by an address.
//
// Prefer this over QueryAccount when you want a balance: an account's Coins field
// holds only the chain's gas denom, and every other denom lives in its own store key.
// Reading the account instead is a mistake that has already been made in this repo.
//
// Costs grow with the number of denoms the address holds, and anyone can send an
// address a new denom without its consent, so treat the cost as caller-influenced.
func (c *Client) QueryBalance(addr crypto.Address) (std.Coins, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("bank/balances/%s", crypto.AddressToBech32(addr))
	qres, err := c.RPCClient.ABCIQuery(context.Background(), path, []byte{})
	if err != nil {
		return nil, nil, errors.Wrap(err, "query balance")
	}
	if qres.Response.Error != nil {
		return nil, nil, errors.Wrap(qres.Response.Error, "query balance")
	}

	var coins std.Coins
	if err := amino.UnmarshalJSON(qres.Response.Data, &coins); err != nil {
		return nil, nil, err
	}
	return coins, qres, nil
}

// QuerySupply retrieves the total supply of a single denomination.
//
// A denomination nobody holds reports zero, which is also what an unknown
// denomination reports; the two are indistinguishable.
func (c *Client) QuerySupply(denom string) (int64, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return 0, nil, err
	}
	if err := std.ValidateDenom(denom); err != nil {
		return 0, nil, errors.Wrap(err, "query supply")
	}

	// The denom is appended whole rather than escaped: a realm-issued denom is
	// "/pkgPath:base" and the route takes the path remainder, so its slashes are
	// part of the denom rather than path separators.
	path := "bank/supply/" + denom
	qres, err := c.RPCClient.ABCIQuery(context.Background(), path, []byte{})
	if err != nil {
		return 0, nil, errors.Wrap(err, "query supply")
	}
	if qres.Response.Error != nil {
		return 0, nil, errors.Wrap(qres.Response.Error, "query supply")
	}

	var supply int64
	if err := amino.UnmarshalJSON(qres.Response.Data, &supply); err != nil {
		return 0, nil, err
	}
	return supply, qres, nil
}

// QuerySessionAccount retrieves session account information for a given masterAddr and sessionAddr.
func (c *Client) QuerySessionAccount(masterAddr, sessionAddr crypto.Address) (*gnoland.GnoSessionAccount, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, nil, err
	}

	path := fmt.Sprintf("auth/accounts/%s/session/%s", crypto.AddressToBech32(masterAddr), crypto.AddressToBech32(sessionAddr))
	data := []byte{}

	qres, err := c.RPCClient.ABCIQuery(context.Background(), path, data)
	if err != nil {
		return nil, nil, errors.Wrap(err, "query session account")
	}
	if len(qres.Response.Data) == 0 || string(qres.Response.Data) == "null" {
		return nil, nil, std.ErrUnknownAddress("unknown master address: " + crypto.AddressToBech32(masterAddr) +
			" or session address: " + crypto.AddressToBech32(sessionAddr))
	}

	qret := &gnoland.GnoSessionAccount{}
	err = amino.UnmarshalJSON(qres.Response.Data, qret)
	if err != nil {
		return nil, nil, err
	}

	return qret, qres, nil
}

// QueryAppVersion retrieves information about the app version
func (c *Client) QueryAppVersion() (string, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return "", nil, err
	}

	path := ".app/version"
	data := []byte{}

	qres, err := c.RPCClient.ABCIQuery(context.Background(), path, data)
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
	data := fmt.Appendf(nil, "%s:%s", pkgPath, args)

	qres, err := c.RPCClient.ABCIQuery(context.Background(), path, data)
	if err != nil {
		return "", nil, errors.Wrap(err, "query render")
	}
	if qres.Response.Error != nil {
		return "", nil, errors.Wrapf(qres.Response.Error, "Render failed: log:%s", qres.Response.Log)
	}

	return string(qres.Response.Data), qres, nil
}

// QEval evaluates the given expression with the realm code at pkgPath. The pkgPath should
// include the prefix like "gno.land/". The expression is usually a function call like
// "GetBoardIDFromName(\"testboard\")". The return value is a typed expression like
// "(1 gno.land/r/archive/boards.BoardID)\n(true bool)".
func (c *Client) QEval(pkgPath string, expression string) (string, *ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return "", nil, err
	}

	path := "vm/qeval"
	data := fmt.Appendf(nil, "%s.%s", pkgPath, expression)

	qres, err := c.RPCClient.ABCIQuery(context.Background(), path, data)
	if err != nil {
		return "", nil, errors.Wrap(err, "query qeval")
	}
	if qres.Response.Error != nil {
		return "", nil, errors.Wrapf(qres.Response.Error, "QEval failed: log:%s", qres.Response.Log)
	}

	return string(qres.Response.Data), qres, nil
}

// Block gets the latest block at height, if any
// Height must be larger than 0
func (c *Client) Block(height int64) (*ctypes.ResultBlock, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, ErrMissingRPCClient
	}

	if height <= 0 {
		return nil, ErrInvalidBlockHeight
	}

	block, err := c.RPCClient.Block(context.Background(), &height)
	if err != nil {
		return nil, fmt.Errorf("block query failed: %w", err)
	}

	return block, nil
}

// BlockResult gets the block results at height, if any
// Height must be larger than 0
func (c *Client) BlockResult(height int64) (*ctypes.ResultBlockResults, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, ErrMissingRPCClient
	}

	if height <= 0 {
		return nil, ErrInvalidBlockHeight
	}

	blockResults, err := c.RPCClient.BlockResults(context.Background(), &height)
	if err != nil {
		return nil, fmt.Errorf("block query failed: %w", err)
	}

	return blockResults, nil
}

// LatestBlockHeight gets the latest block height on the chain
func (c *Client) LatestBlockHeight() (int64, error) {
	if err := c.validateRPCClient(); err != nil {
		return 0, ErrMissingRPCClient
	}

	status, err := c.RPCClient.Status(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("block number query failed: %w", err)
	}

	return status.SyncInfo.LatestBlockHeight, nil
}

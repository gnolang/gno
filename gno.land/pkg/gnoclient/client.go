package gnoclient

import (
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Client provides an interface for interacting with the blockchain.
type Client struct {
	Signer    Signer           // Signer for transaction authentication
	RPCClient rpcclient.Client // RPC client for blockchain communication
}

// Public Client's interface
type IClient interface {
	Query(cfg QueryCfg) (*ctypes.ResultABCIQuery, error)
	QueryAccount(addr crypto.Address) (*std.BaseAccount, *ctypes.ResultABCIQuery, error)
	QueryAppVersion() (string, *ctypes.ResultABCIQuery, error)
	Render(pkgPath string, args string) (string, *ctypes.ResultABCIQuery, error)
	QEval(pkgPath string, expression string) (string, *ctypes.ResultABCIQuery, error)
	Block(height int64) (*ctypes.ResultBlock, error)
	BlockResult(height int64) (*ctypes.ResultBlockResults, error)
	LatestBlockHeight() (int64, error)

	Call(cfg BaseTxCfg, msgs ...MsgCall) (*ctypes.ResultBroadcastTxCommit, error)
	Run(cfg BaseTxCfg, msgs ...MsgRun) (*ctypes.ResultBroadcastTxCommit, error)
	Send(cfg BaseTxCfg, msgs ...MsgSend) (*ctypes.ResultBroadcastTxCommit, error)
	AddPackage(cfg BaseTxCfg, msgs ...MsgAddPackage) (*ctypes.ResultBroadcastTxCommit, error)

	NewSponsorTransaction(cfg SponsorTxCfg, msgs ...Msg) (*std.Tx, error)
	SignTransaction(tx std.Tx, accountNumber, sequenceNumber uint64) (*std.Tx, error)
	ExecuteSponsorTransaction(tx std.Tx, accountNumber, sequenceNumber uint64) (*ctypes.ResultBroadcastTxCommit, error)
}

var _ IClient = (*Client)(nil)

// validateSigner checks that the Client's fields are correctly configured.
func (c *Client) IsValid() error {
	if err := c.validateSigner(); err != nil {
		return err
	}

	if err := c.validateRPCClient(); err != nil {
		return err
	}

	return nil
}

// validateSigner checks that the signer is correctly configured.
func (c *Client) validateSigner() error {
	if c.Signer == nil {
		return ErrMissingSigner
	}

	return nil
}

// validateRPCClient checks that the RPCClient is correctly configured.
func (c *Client) validateRPCClient() error {
	if c.RPCClient == nil {
		return ErrMissingRPCClient
	}

	return nil
}

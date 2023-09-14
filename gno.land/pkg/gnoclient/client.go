package gnoclient

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type Client struct {
	Signer    Signer
	RPCClient rpcclient.Client
}

// validateSigner checks that the signer is correctly configured.
func (c Client) validateSigner() error {
	if c.Signer == nil {
		return errors.New("missing c.Signer")
	}
	return nil
}

// validateRPCClient checks that the RPCClient is correctly configured.
func (c Client) validateRPCClient() error {
	if c.RPCClient == nil {
		return errors.New("missing c.RPCClient")
	}
	return nil
}

type QueryCfg struct {
	Path string
	Data []byte
	client.ABCIQueryOptions
}

// XXX: not sure if we should keep this helper or encourage people to use ABCIQueryWithOptions directly.
func (c Client) Query(cfg QueryCfg) (*ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}
	return c.RPCClient.ABCIQueryWithOptions(cfg.Path, cfg.Data, cfg.ABCIQueryOptions)
}

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

type CallCfg struct {
	PkgPath  string
	FuncName string
	Args     []string

	GasFee         string
	GasWanted      int64
	Send           string
	AccountNumber  uint64
	SequenceNumber uint64
	Memo           string
}

func (c *Client) Call(cfg CallCfg) (*ctypes.ResultBroadcastTxCommit, error) {
	// validate required client fields.
	if err := c.validateSigner(); err != nil {
		return nil, errors.Wrap(err, "validate signer")
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, errors.Wrap(err, "validate RPC client")
	}

	pkgPath := cfg.PkgPath
	funcName := cfg.FuncName
	args := cfg.Args
	gasWanted := cfg.GasWanted
	gasFee := cfg.GasFee
	send := cfg.Send
	sequenceNumber := cfg.SequenceNumber
	accountNumber := cfg.AccountNumber
	memo := cfg.Memo

	// validate config.
	if pkgPath == "" {
		return nil, errors.New("missing PkgPath")
	}
	if funcName == "" {
		return nil, errors.New("missing FuncName")
	}

	// Parse send amount.
	sendCoins, err := std.ParseCoins(send)
	if err != nil {
		return nil, errors.Wrap(err, "parsing send coins")
	}

	// parse gas wanted & fee.
	gasFeeCoins, err := std.ParseCoin(gasFee)
	if err != nil {
		return nil, errors.Wrap(err, "parsing gas fee coin")
	}

	caller := c.Signer.Info().GetAddress()

	// construct msg & tx and marshal.
	msg := vm.MsgCall{
		Caller:  caller,
		Send:    sendCoins,
		PkgPath: pkgPath,
		Func:    funcName,
		Args:    args,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       memo,
	}

	return c.signAndBroadcastTxCommit(tx, accountNumber, sequenceNumber)
}

func (c Client) signAndBroadcastTxCommit(tx std.Tx, accountNumber, sequenceNumber uint64) (*ctypes.ResultBroadcastTxCommit, error) {
	caller := c.Signer.Info().GetAddress()

	if sequenceNumber == 0 || accountNumber == 0 {
		account, _, err := c.QueryAccount(caller.String())
		if err != nil {
			return nil, errors.Wrap(err, "query account")
		}
		accountNumber = account.AccountNumber
		sequenceNumber = account.Sequence
	}

	signCfg := SignCfg{
		UnsignedTX:     tx,
		SequenceNumber: sequenceNumber,
		AccountNumber:  accountNumber,
	}
	signedTx, err := c.Signer.Sign(signCfg)
	if err != nil {
		return nil, errors.Wrap(err, "sign")
	}

	bz, err := amino.Marshal(signedTx)
	if err != nil {
		return nil, errors.Wrap(err, "remarshaling tx binary bytes")
	}

	bres, err := c.RPCClient.BroadcastTxCommit(bz)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting bytes")
	}

	if bres.CheckTx.IsErr() {
		return nil, errors.Wrap(bres.CheckTx.Error, "check transaction failed: log:%s", bres.CheckTx.Log)
	}
	if bres.DeliverTx.IsErr() {
		return nil, errors.Wrap(bres.DeliverTx.Error, "deliver transaction failed: log:%s", bres.DeliverTx.Log)
	}

	return bres, nil
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

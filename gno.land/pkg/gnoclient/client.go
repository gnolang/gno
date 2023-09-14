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

	GasFee    string
	GasWanted int64
	Send      string

	AccountNumber  uint64
	SequenceNumber uint64
}

func (c *Client) Call(cfg CallCfg) error {
	// validate config.
	if cfg.PkgPath == "" {
		return errors.New("missing PkgPath")
	}
	if cfg.FuncName == "" {
		return errors.New("missing FuncName")
	}

	// Parse send amount.
	sendCoins, err := std.ParseCoins(cfg.Send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}

	// parse gas wanted & fee.
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// validate required client fields.
	if err := c.validateSigner(); err != nil {
		return err
	}
	if err := c.validateRPCClient(); err != nil {
		return err
	}

	caller := c.Signer.Info().GetAddress()

	// construct msg & tx and marshal.
	msg := vm.MsgCall{
		Caller:  caller,
		Send:    sendCoins,
		PkgPath: cfg.PkgPath,
		Func:    cfg.FuncName,
		Args:    cfg.Args,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       "",
	}

	if cfg.SequenceNumber == 0 || cfg.AccountNumber == 0 {
		account, _, err := c.QueryAccount(caller.String())
		if err != nil {
			return errors.Wrap(err, "query account")
		}
		cfg.AccountNumber = account.AccountNumber
		cfg.SequenceNumber = account.Sequence
	}

	signCfg := SignCfg{
		UnsignedTX:     tx,
		SequenceNumber: cfg.SequenceNumber,
		AccountNumber:  cfg.AccountNumber,
	}
	signedTx, err := c.Signer.Sign(signCfg)
	if err != nil {
		return errors.Wrap(err, "sign")
	}

	_ = signedTx
	/*
		// broadcast signed tx
		bopts := &broadcastCfg{
			remote: c.remote,
			tx:     signedTx,
		}
		bres, err := broadcastHandler(bopts)
		if err != nil {
			return errors.Wrap(err, "broadcast tx")
		}
		if bres.CheckTx.IsErr() {
			return errors.Wrap(bres.CheckTx.Error, "check transaction failed: log:%s", bres.CheckTx.Log)
		}
		if bres.DeliverTx.IsErr() {
			return errors.Wrap(bres.DeliverTx.Error, "deliver transaction failed: log:%s", bres.DeliverTx.Log)
		}
	*/

	return nil
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

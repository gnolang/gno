package gnoclient

import (
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
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

func (c Client) Query(cfg QueryCfg) (*ctypes.ResultABCIQuery, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}
	return c.RPCClient.ABCIQueryWithOptions(cfg.Path, cfg.Data, cfg.ABCIQueryOptions)
}

/*
func (c *Client) Call(
	pkgPath string,
	fnc string,
	args []string,
	gasFee string,
	gasWanted int64,
	send string,
) error {
	if err := c.validateSigner(); err != nil {
		return err
	}

	caller := c.Signer.Info().GetAddress()

	// Parse send amount.
	sendCoins, err := std.ParseCoins(send)
	if err != nil {
		return errors.Wrap(err, "parsing send coins")
	}

	// parse gas wanted & fee.
	gasFeeCoins, err := std.ParseCoin(gasFee)
	if err != nil {
		return errors.Wrap(err, "parsing gas fee coin")
	}

	// construct msg & tx and marshal.
	msg := vm.MsgCall{
		Caller:  caller,
		Send:    sendCoins,
		PkgPath: pkgPath,
		Func:    fnc,
		Args:    args,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       "",
	}

	qopts := &queryCfg{
		remote: c.remote,
		path:   fmt.Sprintf("auth/accounts/%s", caller),
	}
	qres, err := queryHandler(qopts)
	if err != nil {
		return errors.Wrap(err, "query account")
	}
	var qret struct{ BaseAccount std.BaseAccount }
	err = amino.UnmarshalJSON(qres.Response.Data, &qret)
	if err != nil {
		return err
	}

	// sign tx
	accountNumber := qret.BaseAccount.AccountNumber
	sequence := qret.BaseAccount.Sequence
	sopts := &signCfg{
		kb:            c.keybase,
		sequence:      sequence,
		accountNumber: accountNumber,
		chainID:       c.chainID,
		nameOrBech32:  nameOrBech32,
		txJSON:        amino.MustMarshalJSON(tx),
		pass:          password,
	}

	signedTx, err := SignHandler(sopts)
	if err != nil {
		return errors.Wrap(err, "sign tx")
	}

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

	return nil
}
*/

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

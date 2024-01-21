package gnoclient

import (
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// MsgCall - syntax sugar for vm.MsgCall
type MsgCall struct {
	PkgPath  string   // Package path
	FuncName string   // Function name
	Args     []string // Function arguments
	Send     string   // Send amount
}

// CallCfg contains configuration options for executing a contract call.
type CallCfg struct {
	MsgCall
	GasFee         string // Gas fee
	GasWanted      int64  // Gas wanted
	AccountNumber  uint64 // Account number
	SequenceNumber uint64 // Sequence number
	Memo           string // Memo
}

// MultiCallCfg contains configuration options for executing a contract call.
type MultiCallCfg struct {
	Msgs []MsgCall

	// Per Tx
	GasFee         string // Gas fee
	GasWanted      int64  // Gas wanted
	AccountNumber  uint64 // Account number
	SequenceNumber uint64 // Sequence number
	Memo           string // Memo
}

// Call executes a contract call on the blockchain.
func (c *Client) Call(cfg CallCfg) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
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

	// Validate config.
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

	// Parse gas wanted & fee.
	gasFeeCoins, err := std.ParseCoin(gasFee)
	if err != nil {
		return nil, errors.Wrap(err, "parsing gas fee coin")
	}

	caller := c.Signer.Info().GetAddress()

	// Construct message & transaction and marshal.
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

// MultiCall executes a contract call on the blockchain.
func (c *Client) MultiCall(cfg MultiCallCfg) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validateSigner(); err != nil {
		return nil, errors.Wrap(err, "validate signer")
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, errors.Wrap(err, "validate RPC client")
	}

	sequenceNumber := cfg.SequenceNumber
	accountNumber := cfg.AccountNumber

	msgs := make([]vm.MsgCall, 0, len(cfg.Msgs))
	for _, msg := range cfg.Msgs {
		pkgPath := msg.PkgPath
		funcName := msg.FuncName
		args := msg.Args
		send := msg.Send

		// Validate config.
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

		if err != nil {
			return nil, errors.Wrap(err, "parsing gas fee coin")
		}

		caller := c.Signer.Info().GetAddress()

		msgs = append(msgs, vm.MsgCall{
			Caller:  caller,
			Send:    sendCoins,
			PkgPath: pkgPath,
			Func:    funcName,
			Args:    args,
		})
	}

	// Cast vm.MsgCall back into std.Msg
	stdMsgs := make([]std.Msg, len(msgs))
	for i, msg := range msgs {
		stdMsgs[i] = msg
	}

	// Parse gas wanted & fee.
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return nil, errors.Wrap(err, "parsing gas fee coin")
	}

	// Pack transaction
	tx := std.Tx{
		Msgs:       stdMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       "",
	}

	return c.signAndBroadcastTxCommit(tx, accountNumber, sequenceNumber)
}

// signAndBroadcastTxCommit signs a transaction and broadcasts it, returning the result.
func (c Client) signAndBroadcastTxCommit(tx std.Tx, accountNumber, sequenceNumber uint64) (*ctypes.ResultBroadcastTxCommit, error) {
	caller := c.Signer.Info().GetAddress()

	if sequenceNumber == 0 || accountNumber == 0 {
		account, _, err := c.QueryAccount(caller)
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
		return nil, errors.Wrap(err, "marshaling tx binary bytes")
	}

	bres, err := c.RPCClient.BroadcastTxCommit(bz)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting bytes")
	}

	if bres.CheckTx.IsErr() {
		return bres, errors.Wrap(bres.CheckTx.Error, "check transaction failed: log:%s", bres.CheckTx.Log)
	}
	if bres.DeliverTx.IsErr() {
		return bres, errors.Wrap(bres.DeliverTx.Error, "deliver transaction failed: log:%s", bres.DeliverTx.Log)
	}

	return bres, nil
}

// TODO: Add more functionality, examples, and unit tests.

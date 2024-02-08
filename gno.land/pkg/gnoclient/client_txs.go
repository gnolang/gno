package gnoclient

import (
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	ErrEmptyPkgPath     = errors.New("empty pkg path")
	ErrEmptyFuncName    = errors.New("empty function name")
	ErrEmptyPackage     = errors.New("empty package to run")
	ErrInvalidGasWanted = errors.New("invalid gas wanted")
	ErrInvalidGasFee    = errors.New("invalid gas fee")
	ErrMissingSigner    = errors.New("missing Signer")
	ErrMissingRPCClient = errors.New("missing RPCClient")
)

type BaseTxCfg struct {
	GasFee         string // Gas fee
	GasWanted      int64  // Gas wanted
	AccountNumber  uint64 // Account number
	SequenceNumber uint64 // Sequence number
	Memo           string // Memo
}

// MsgCall - syntax sugar for vm.MsgCall
type MsgCall struct {
	PkgPath  string   // Package path
	FuncName string   // Function name
	Args     []string // Function arguments
	Send     string   // Send amount
}

type MsgRun struct {
	Package *std.MemPackage // Package to run
	Send    string          // Send amount
}

// Call executes a contract call on the blockchain.
func (c *Client) Call(cfg BaseTxCfg, msgs ...MsgCall) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validateSigner(); err != nil {
		return nil, err
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	// Validate base transaction config
	if err := cfg.validateBaseTxConfig(); err != nil {
		return nil, err
	}

	// Parse MsgCall slice
	vmMsgs := make([]std.Msg, 0, len(msgs))
	for _, msg := range msgs {
		// Validate MsgCall fields
		if err := msg.validateMsgCall(); err != nil {
			return nil, err
		}

		// Parse send coins
		send, err := std.ParseCoins(msg.Send)
		if err != nil {
			return nil, err
		}

		// Unwrap syntax sugar to vm.MsgCall slice
		vmMsgs = append(vmMsgs, std.Msg(vm.MsgCall{
			Caller:  c.Signer.Info().GetAddress(),
			PkgPath: msg.PkgPath,
			Func:    msg.FuncName,
			Args:    msg.Args,
			Send:    send,
		}))
	}

	// Parse gas fee
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return nil, err
	}

	// Pack transaction
	tx := std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}

	return c.signAndBroadcastTxCommit(tx, cfg.AccountNumber, cfg.SequenceNumber)
}

func (c *Client) Run(cfg BaseTxCfg, msgs ...MsgRun) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields
	if err := c.validateSigner(); err != nil {
		return nil, err
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	// Validate base transaction config
	if err := cfg.validateBaseTxConfig(); err != nil {
		return nil, err
	}

	// Parse MsgRun slice
	vmMsgs := make([]std.Msg, 0, len(msgs))
	for _, msg := range msgs {
		// Validate MsgCall fields
		if err := msg.validateMsgRun(); err != nil {
			return nil, err
		}

		// Parse send coins
		send, err := std.ParseCoins(msg.Send)
		if err != nil {
			return nil, err
		}

		caller := c.Signer.Info().GetAddress()

		// Precompile and validate Gno syntax
		if err = gno.PrecompileAndCheckMempkg(msg.Package); err != nil {
			return nil, err
		}

		msg.Package.Name = "main"
		msg.Package.Path = "gno.land/r/" + caller.String() + "/run"

		// Unwrap syntax sugar to vm.MsgCall slice
		vmMsgs = append(vmMsgs, std.Msg(vm.MsgRun{
			Caller:  caller,
			Package: msg.Package,
			Send:    send,
		}))
	}

	// Parse gas fee
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return nil, err
	}

	// Pack transaction
	tx := std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}

	return c.signAndBroadcastTxCommit(tx, cfg.AccountNumber, cfg.SequenceNumber)
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

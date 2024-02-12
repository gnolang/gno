package gnoclient

import (
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	ErrEmptyPkgPath      = errors.New("empty pkg path")
	ErrEmptyFuncName     = errors.New("empty function name")
	ErrInvalidGasWanted  = errors.New("invalid gas wanted")
	ErrInvalidGasFee     = errors.New("invalid gas fee")
	ErrMissingSigner     = errors.New("missing Signer")
	ErrMissingRPCClient  = errors.New("missing RPCClient")
	ErrInvalidToAddress  = errors.New("invalid send to address")
	ErrInvalidSendAmount = errors.New("invalid send amount")
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

// MsgSend - syntax sugar for bank.MsgSend minus fields in BaseTxCfg
type MsgSend struct {
	ToAddress crypto.Address // Send to address
	Send      string         // Send amount
}

// RunCfg contains configuration options for running a temporary package on the blockchain.
type RunCfg struct {
	Package        *std.MemPackage
	GasFee         string // Gas fee
	GasWanted      int64  // Gas wanted
	AccountNumber  uint64 // Account number
	SequenceNumber uint64 // Sequence number
	Memo           string // Memo
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
	vmMsgs := make([]vm.MsgCall, 0, len(msgs))
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
		vmMsgs = append(vmMsgs, vm.MsgCall{
			Caller:  c.Signer.Info().GetAddress(),
			PkgPath: msg.PkgPath,
			Func:    msg.FuncName,
			Args:    msg.Args,
			Send:    send,
		})
	}

	// Cast vm.MsgCall back into std.Msg
	stdMsgs := make([]std.Msg, len(vmMsgs))
	for i, msg := range vmMsgs {
		stdMsgs[i] = msg
	}

	// Parse gas fee
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return nil, err
	}

	// Pack transaction
	tx := std.Tx{
		Msgs:       stdMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}

	return c.signAndBroadcastTxCommit(tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// Send currency to an account on the blockchain.
func (c *Client) Send(cfg BaseTxCfg, msgs ...MsgSend) (*ctypes.ResultBroadcastTxCommit, error) {
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

	// Parse MsgSend slice
	vmMsgs := make([]bank.MsgSend, 0, len(msgs))
	for _, msg := range msgs {
		// Validate MsgSend fields
		if err := msg.validateMsgSend(); err != nil {
			return nil, err
		}

		// Parse send coins
		send, err := std.ParseCoins(msg.Send)
		if err != nil {
			return nil, err
		}

		// Unwrap syntax sugar to vm.MsgSend slice
		vmMsgs = append(vmMsgs, bank.MsgSend{
			FromAddress: c.Signer.Info().GetAddress(),
			ToAddress:   msg.ToAddress,
			Amount:      send,
		})
	}

	// Cast vm.MsgSend back into std.Msg
	stdMsgs := make([]std.Msg, len(vmMsgs))
	for i, msg := range vmMsgs {
		stdMsgs[i] = msg
	}

	// Parse gas fee
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return nil, err
	}

	// Pack transaction
	tx := std.Tx{
		Msgs:       stdMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}

	return c.signAndBroadcastTxCommit(tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// Temporarily load cfg.Package on the blockchain and run main() which can
// call realm functions and use println() to output to the "console".
// This returns bres where string(bres.DeliverTx.Data) is the "console" output.
func (c *Client) Run(cfg RunCfg) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validateSigner(); err != nil {
		return nil, errors.Wrap(err, "validate signer")
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, errors.Wrap(err, "validate RPC client")
	}

	memPkg := cfg.Package
	gasWanted := cfg.GasWanted
	gasFee := cfg.GasFee
	sequenceNumber := cfg.SequenceNumber
	accountNumber := cfg.AccountNumber
	memo := cfg.Memo

	// Validate config.
	if memPkg.IsEmpty() {
		return nil, errors.New("found an empty package " + memPkg.Path)
	}

	// Parse gas wanted & fee.
	gasFeeCoins, err := std.ParseCoin(gasFee)
	if err != nil {
		return nil, errors.Wrap(err, "parsing gas fee coin")
	}

	caller := c.Signer.Info().GetAddress()

	// precompile and validate syntax
	err = gno.PrecompileAndCheckMempkg(memPkg)
	if err != nil {
		return nil, errors.Wrap(err, "precompile and check")
	}
	memPkg.Name = "main"
	memPkg.Path = ""

	// Construct message & transaction and marshal.
	msg := vm.MsgRun{
		Caller:  caller,
		Package: memPkg,
	}
	tx := std.Tx{
		Msgs:       []std.Msg{msg},
		Fee:        std.NewFee(gasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       memo,
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

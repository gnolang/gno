package gnoclient

import (
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	ErrInvalidGasWanted = errors.New("invalid gas wanted")
	ErrInvalidGasFee    = errors.New("invalid gas fee")
	ErrMissingSigner    = errors.New("missing Signer")
	ErrMissingRPCClient = errors.New("missing RPCClient")
)

// BaseTxCfg defines the base transaction configuration, shared by all message types
type BaseTxCfg struct {
	GasFee         string         // Gas fee
	GasWanted      int64          // Gas wanted
	AccountNumber  uint64         // Account number
	SequenceNumber uint64         // Sequence number
	Memo           string         // Memo
	CallerAddress  crypto.Address // The caller Address if known
}

// Call executes one or more MsgCall calls on the blockchain
func (c *Client) Call(cfg BaseTxCfg, msgs ...vm.MsgCall) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	// MakeCallTx calls validateSigner().
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	tx, err := c.MakeCallTx(cfg, msgs...)
	if err != nil {
		return nil, err
	}
	return c.signAndBroadcastTxCommit(*tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// MakeCallTx makes an unsigned transaction from one or more MsgCall.
// If cfg.CallerAddress.IsZero() then get it from c.Signer.
func (c *Client) MakeCallTx(cfg BaseTxCfg, msgs ...MsgCall) (*std.Tx, error) {
	if cfg.CallerAddress.IsZero() {
		// Validate required client fields
		if err := c.validateSigner(); err != nil {
			return nil, err
		}
	}

	// Validate base transaction config
	if err := cfg.validateBaseTxConfig(); err != nil {
		return nil, err
	}

	vmMsgs := make([]std.Msg, 0, len(msgs))
	for _, msg := range msgs {
		// Validate MsgCall fields
		if err := msg.ValidateBasic(); err != nil {
			return nil, err
		}

		vmMsgs = append(vmMsgs, std.Msg(msg))
	}

	// Parse gas fee
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return nil, err
	}

	// Pack transaction
	return &std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}, nil
}

// Run executes one or more MsgRun calls on the blockchain
func (c *Client) Run(cfg BaseTxCfg, msgs ...vm.MsgRun) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	// MakeRunTx calls validateSigner().
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	tx, err := c.MakeRunTx(cfg, msgs...)
	if err != nil {
		return nil, err
	}
	return c.signAndBroadcastTxCommit(*tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// MakeRunTx makes an unsigned transaction from one or more MsgRun.
// If cfg.CallerAddress.IsZero() then get it from c.Signer.
func (c *Client) MakeRunTx(cfg BaseTxCfg, msgs ...MsgRun) (*std.Tx, error) {
	if cfg.CallerAddress.IsZero() {
		// Validate required client fields
		if err := c.validateSigner(); err != nil {
			return nil, err
		}
	}

	// Validate base transaction config
	if err := cfg.validateBaseTxConfig(); err != nil {
		return nil, err
	}

	vmMsgs := make([]std.Msg, 0, len(msgs))
	for _, msg := range msgs {
		// Validate MsgRun fields
		if err := msg.ValidateBasic(); err != nil {
			return nil, err
		}

		vmMsgs = append(vmMsgs, std.Msg(msg))
	}

	// Parse gas fee
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return nil, err
	}

	// Pack transaction
	return &std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}, nil
}

// Send executes one or more MsgSend calls on the blockchain
func (c *Client) Send(cfg BaseTxCfg, msgs ...bank.MsgSend) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	// MakeSendTx calls validateSigner().
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	tx, err := c.MakeSendTx(cfg, msgs...)
	if err != nil {
		return nil, err
	}
	return c.signAndBroadcastTxCommit(*tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// MakeSendTx makes an unsigned transaction from one or more MsgSend.
// If cfg.CallerAddress.IsZero() then get it from c.Signer.
func (c *Client) MakeSendTx(cfg BaseTxCfg, msgs ...MsgSend) (*std.Tx, error) {
	if cfg.CallerAddress.IsZero() {
		// Validate required client fields
		if err := c.validateSigner(); err != nil {
			return nil, err
		}
	}

	// Validate base transaction config
	if err := cfg.validateBaseTxConfig(); err != nil {
		return nil, err
	}

	vmMsgs := make([]std.Msg, 0, len(msgs))
	for _, msg := range msgs {
		// Validate MsgSend fields
		if err := msg.ValidateBasic(); err != nil {
			return nil, err
		}

		vmMsgs = append(vmMsgs, std.Msg(msg))
	}

	// Parse gas fee
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return nil, err
	}

	// Pack transaction
	return &std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}, nil
}

// AddPackage executes one or more AddPackage calls on the blockchain
func (c *Client) AddPackage(cfg BaseTxCfg, msgs ...vm.MsgAddPackage) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	// MakeAddPackageTx calls validateSigner().
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	tx, err := c.MakeAddPackageTx(cfg, msgs...)
	if err != nil {
		return nil, err
	}
	return c.signAndBroadcastTxCommit(*tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// MakeAddPackageTx makes an unsigned transaction from one or more MsgAddPackage.
// If cfg.CallerAddress.IsZero() then get it from c.Signer.
func (c *Client) MakeAddPackageTx(cfg BaseTxCfg, msgs ...MsgAddPackage) (*std.Tx, error) {
	if cfg.CallerAddress.IsZero() {
		// Validate required client fields
		if err := c.validateSigner(); err != nil {
			return nil, err
		}
	}

	// Validate base transaction config
	if err := cfg.validateBaseTxConfig(); err != nil {
		return nil, err
	}

	vmMsgs := make([]std.Msg, 0, len(msgs))
	for _, msg := range msgs {
		// Validate MsgAddPackage fields
		if err := msg.ValidateBasic(); err != nil {
			return nil, err
		}

		vmMsgs = append(vmMsgs, std.Msg(msg))
	}

	// Parse gas fee
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return nil, err
	}

	// Pack transaction
	return &std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}, nil
}

// signAndBroadcastTxCommit signs a transaction and broadcasts it, returning the result
func (c *Client) signAndBroadcastTxCommit(tx std.Tx, accountNumber, sequenceNumber uint64) (*ctypes.ResultBroadcastTxCommit, error) {
	signedTx, err := c.SignTx(tx, accountNumber, sequenceNumber)
	if err != nil {
		return nil, err
	}
	return c.BroadcastTxCommit(signedTx)
}

// SignTx signs a transaction and returns a signed tx ready for broadcasting.
// If accountNumber or sequenceNumber is 0 then query the blockchain for the value.
func (c *Client) SignTx(tx std.Tx, accountNumber, sequenceNumber uint64) (*std.Tx, error) {
	if err := c.validateSigner(); err != nil {
		return nil, err
	}
	caller, err := c.Signer.Info()
	if err != nil {
		return nil, err
	}

	if sequenceNumber == 0 || accountNumber == 0 {
		account, _, err := c.QueryAccount(caller.GetAddress())
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
	return signedTx, nil
}

// BroadcastTxCommit marshals and broadcasts the signed transaction, returning the result.
// If the result has a delivery error, then return a wrapped error.
func (c *Client) BroadcastTxCommit(signedTx *std.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, err
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

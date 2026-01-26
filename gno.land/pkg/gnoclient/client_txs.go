package gnoclient

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mempool"
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

const simulatePath = ".app/simulate"

// BaseTxCfg defines the base transaction configuration, shared by all message types
type BaseTxCfg struct {
	GasFee         string // Gas fee
	GasWanted      int64  // Gas wanted
	AccountNumber  uint64 // Account number
	SequenceNumber uint64 // Sequence number
	Memo           string // Memo
}

// Call executes one or more MsgCall calls on the blockchain
func (c *Client) Call(cfg BaseTxCfg, msgs ...vm.MsgCall) (*mempool.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validateSigner(); err != nil {
		return nil, err
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	tx, err := NewCallTx(cfg, msgs...)
	if err != nil {
		return nil, err
	}
	return c.signAndBroadcastTxCommit(*tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// NewCallTx makes an unsigned transaction from one or more MsgCall.
// The Caller field must be set.
func NewCallTx(cfg BaseTxCfg, msgs ...vm.MsgCall) (*std.Tx, error) {
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
func (c *Client) Run(cfg BaseTxCfg, msgs ...vm.MsgRun) (*mempool.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validateSigner(); err != nil {
		return nil, err
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	tx, err := NewRunTx(cfg, msgs...)
	if err != nil {
		return nil, err
	}
	return c.signAndBroadcastTxCommit(*tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// NewRunTx makes an unsigned transaction from one or more MsgRun.
// The Caller field must be set.
func NewRunTx(cfg BaseTxCfg, msgs ...vm.MsgRun) (*std.Tx, error) {
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
func (c *Client) Send(cfg BaseTxCfg, msgs ...bank.MsgSend) (*mempool.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validateSigner(); err != nil {
		return nil, err
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	tx, err := NewSendTx(cfg, msgs...)
	if err != nil {
		return nil, err
	}
	return c.signAndBroadcastTxCommit(*tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// NewSendTx makes an unsigned transaction from one or more MsgSend.
// The FromAddress field must be set.
func NewSendTx(cfg BaseTxCfg, msgs ...bank.MsgSend) (*std.Tx, error) {
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
func (c *Client) AddPackage(cfg BaseTxCfg, msgs ...vm.MsgAddPackage) (*mempool.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validateSigner(); err != nil {
		return nil, err
	}
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	tx, err := NewAddPackageTx(cfg, msgs...)
	if err != nil {
		return nil, err
	}
	return c.signAndBroadcastTxCommit(*tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// NewAddPackageTx makes an unsigned transaction from one or more MsgAddPackage.
// The Creator field must be set.
func NewAddPackageTx(cfg BaseTxCfg, msgs ...vm.MsgAddPackage) (*std.Tx, error) {
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
func (c *Client) signAndBroadcastTxCommit(tx std.Tx, accountNumber, sequenceNumber uint64) (*mempool.ResultBroadcastTxCommit, error) {
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
func (c *Client) BroadcastTxCommit(signedTx *std.Tx) (*mempool.ResultBroadcastTxCommit, error) {
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}
	bz, err := amino.Marshal(signedTx)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling tx binary bytes")
	}

	bres, err := c.RPCClient.BroadcastTxCommit(context.Background(), bz)
	if err != nil {
		return nil, errors.Wrap(err, "broadcasting bytes")
	}

	if bres.CheckTx.IsErr() {
		return bres, errors.Wrapf(bres.CheckTx.Error, "check transaction failed: log:%s", bres.CheckTx.Log)
	}
	if bres.DeliverTx.IsErr() {
		return bres, errors.Wrapf(bres.DeliverTx.Error, "deliver transaction failed: log:%s", bres.DeliverTx.Log)
	}

	return bres, nil
}

// EstimateGas returns the least amount of gas required
// for the transaction to go through on the chain (minimum gas wanted).
// The estimation process assumes the transaction signature has the proper public key
func (c *Client) EstimateGas(tx *std.Tx) (int64, error) {
	deliverTx, err := c.Simulate(tx)
	if err != nil {
		return 0, err
	}

	// Return the actual value returned by the node
	// for executing the transaction
	return deliverTx.GasUsed, nil
}

// Simulate the transaction and return the ResponseDeliverTx.
// The simulation process assumes the transaction signature has the proper public key
func (c *Client) Simulate(tx *std.Tx) (*abci.ResponseDeliverTx, error) {
	// Make sure the RPC client is set
	if err := c.validateRPCClient(); err != nil {
		return nil, err
	}

	// Prepare the transaction.
	// The transaction needs to be amino-binary encoded
	// in order to be estimated
	encodedTx, err := amino.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal tx: %w", err)
	}

	// Perform the simulation query
	resp, err := c.RPCClient.ABCIQuery(context.Background(), simulatePath, encodedTx)
	if err != nil {
		return nil, fmt.Errorf("unable to perform ABCI query: %w", err)
	}

	// Extract the query response
	if err = resp.Response.Error; err != nil {
		return nil, fmt.Errorf("error encountered during ABCI query: %w", err)
	}

	deliverTx := new(abci.ResponseDeliverTx)
	if err = amino.Unmarshal(resp.Response.Value, deliverTx); err != nil {
		return nil, fmt.Errorf("unable to unmarshal simulation response: %w", err)
	}

	if err = deliverTx.Error; err != nil {
		return nil, fmt.Errorf("error encountered during simulation: %w", err)
	}

	return deliverTx, nil
}

package gnoclient

import (
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Call executes one or more MsgCall calls on the blockchain
func (c *Client) Call(cfg BaseTxCfg, msgs ...vm.MsgCall) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validate(); err != nil {
		return nil, err
	}

	// Validate base transaction config
	if err := cfg.validate(); err != nil {
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
	tx := std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}

	return c.signAndBroadcastTxCommit(tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// Run executes one or more MsgRun calls on the blockchain
func (c *Client) Run(cfg BaseTxCfg, msgs ...vm.MsgRun) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validate(); err != nil {
		return nil, err
	}

	// Validate base transaction config
	if err := cfg.validate(); err != nil {
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
	tx := std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}

	return c.signAndBroadcastTxCommit(tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// Send executes one or more MsgSend calls on the blockchain
func (c *Client) Send(cfg BaseTxCfg, msgs ...bank.MsgSend) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validate(); err != nil {
		return nil, err
	}

	// Validate base transaction config
	if err := cfg.validate(); err != nil {
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
	tx := std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}

	return c.signAndBroadcastTxCommit(tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// AddPackage executes one or more AddPackage calls on the blockchain
func (c *Client) AddPackage(cfg BaseTxCfg, msgs ...vm.MsgAddPackage) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields.
	if err := c.validate(); err != nil {
		return nil, err
	}

	// Validate base transaction config
	if err := cfg.validate(); err != nil {
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
	tx := std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}

	return c.signAndBroadcastTxCommit(tx, cfg.AccountNumber, cfg.SequenceNumber)
}

// CreateTx creates an signed transaction for various types of messages which used for sponsorship
func (c *Client) NewSponsorTransaction(cfg SponsorTxCfg, msgs ...std.Msg) (*std.Tx, error) {
	// validate required client fields
	if err := c.validate(); err != nil {
		return nil, err
	}

	// Validate sponsor transaction config
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Ensure at least one message is provided
	if len(msgs) == 0 {
		return nil, ErrNoMessages
	}

	// Determine the type of the first user-provided message
	firstMsgType := msgs[0].Type()

	vmMsgs := make([]std.Msg, 0, len(msgs)+1)

	// First msg in list must be MsgNoop
	vmMsgs = append(vmMsgs, vm.MsgNoop{
		Caller: cfg.SponsorAddress,
	})

	for _, msg := range msgs {
		// Check if all messages are of the same type
		if msg.Type() != firstMsgType {
			return nil, ErrMixedMessageTypes
		}

		if err := msg.ValidateBasic(); err != nil {
			return nil, err
		}

		vmMsgs = append(vmMsgs, msg)
	}

	// Parse gas fee
	gasFeeCoins, err := std.ParseCoin(cfg.GasFee)
	if err != nil {
		return nil, err
	}

	// Pack transaction
	tx := &std.Tx{
		Msgs:       vmMsgs,
		Fee:        std.NewFee(cfg.GasWanted, gasFeeCoins),
		Signatures: nil,
		Memo:       cfg.Memo,
	}

	return tx, nil
}

// SignTx signs a transaction using the client's signer
func (c *Client) SignTransaction(tx std.Tx, accountNumber, sequenceNumber uint64) (*std.Tx, error) {
	// Ensure sequence number and account number are provided
	signCfg := SignCfg{
		Tx:             tx,
		SequenceNumber: sequenceNumber,
		AccountNumber:  accountNumber,
	}

	signedTx, err := c.Signer.Sign(signCfg)
	if err != nil {
		return nil, errors.Wrap(err, "sign")
	}

	return signedTx, nil
}

// ExecuteSponsorTransaction allows broadcasting a pre-signed transaction (represented by `sponsorTx`)
// using the signer's account to pay transaction fees. The `sponsoree` account who signed `the sponsorTx“ before benefits
// from this transaction without incurring any gas costs
func (c *Client) ExecuteSponsorTransaction(tx std.Tx, accountNumber, sequenceNumber uint64) (*ctypes.ResultBroadcastTxCommit, error) {
	// Validate required client fields
	if err := c.validate(); err != nil {
		return nil, err
	}

	// Validate basic transaction
	if err := tx.ValidateBasic(); err != nil {
		return nil, err
	}

	// Ensure tx is a sponsor transaction
	if !tx.IsSponsorTx() {
		return nil, ErrInvalidSponsorTx
	}

	return c.signAndBroadcastTxCommit(tx, accountNumber, sequenceNumber)
}

// signAndBroadcastTxCommit signs a transaction and broadcasts it, returning the result
func (c *Client) signAndBroadcastTxCommit(tx std.Tx, accountNumber, sequenceNumber uint64) (*ctypes.ResultBroadcastTxCommit, error) {
	signedTx, err := c.SignTransaction(tx, accountNumber, sequenceNumber)
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

package gnoclient

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Define various error messages for different validation failures
var (
	ErrEmptyPackage      = errors.New("empty package to run")
	ErrEmptyPkgPath      = errors.New("empty pkg path")
	ErrEmptyFuncName     = errors.New("empty function name")
	ErrInvalidGasWanted  = errors.New("invalid gas wanted")
	ErrInvalidGasFee     = errors.New("invalid gas fee")
	ErrMissingSigner     = errors.New("missing Signer")
	ErrMissingRPCClient  = errors.New("missing RPCClient")
	ErrInvalidToAddress  = errors.New("invalid send to address")
	ErrInvalidAmount     = errors.New("invalid send/deposit amount")
	ErrInvalidMsgType    = errors.New("invalid msg type")
	ErrNoMessages        = errors.New("no messages provided")
	ErrMixedMessageTypes = errors.New("mixed message types not allowed")
	ErrNoSignatures      = errors.New("no signatures provided")

	ErrInvalidSponsorAddress = errors.New("invalid sponsor address")
	ErrInvalidSponsorTx      = errors.New("invalid sponsor tx")
)

// Constants for different message types.
const (
	MSG_CALL    = "call"
	MSG_RUN     = "run"
	MSG_SEND    = "send"
	MSG_ADD_PKG = "add_pkg"
)

// Msg defines the interface for different types of messages.
type Msg interface {
	IsValid() error               // Validates the message.
	GetCoins() (std.Coins, error) // Retrieves the coins involved in the message.
	GetType() string              // Returns the type of the message.
}

// BaseTxCfg defines the base transaction configuration shared by all message types.
type BaseTxCfg struct {
	GasFee         string // Gas fee
	GasWanted      int64  // Gas wanted
	AccountNumber  uint64 // Account number
	SequenceNumber uint64 // Sequence number
	Memo           string // Memo
}

// validateBaseTxConfig validates the base transaction configuration.
func (cfg BaseTxCfg) IsValid() error {
	if cfg.GasWanted <= 0 {
		return ErrInvalidGasWanted
	}
	if cfg.GasFee == "" {
		return ErrInvalidGasFee
	}
	return nil
}

type SponsorTxCfg struct {
	BaseTxCfg
	SponsorAddress crypto.Address
}

// validateBaseTxConfig validates the base transaction configuration.
func (cfg SponsorTxCfg) IsValid() error {
	if cfg.SponsorAddress.IsZero() {
		return ErrInvalidSponsorAddress
	}
	if cfg.GasWanted <= 0 {
		return ErrInvalidGasWanted
	}
	if cfg.GasFee == "" {
		return ErrInvalidGasFee
	}
	return nil
}

// MsgCall represents a call message in the VM.
type MsgCall struct {
	PkgPath  string   // Package path
	FuncName string   // Function name
	Args     []string // Function arguments
	Send     string   // Send amount
}

// getType returns the type of the MsgCall.
func (msg MsgCall) GetType() string {
	return MSG_CALL
}

// validateMsg validates the MsgCall.
func (msg MsgCall) IsValid() error {
	if msg.PkgPath == "" {
		return ErrEmptyPkgPath
	}
	if msg.FuncName == "" {
		return ErrEmptyFuncName
	}
	return nil
}

// getCoins retrieves the coins involved in the MsgCall.
func (msg MsgCall) GetCoins() (std.Coins, error) {
	coins, err := std.ParseCoins(msg.Send)
	if err != nil {
		return nil, ErrInvalidAmount
	}
	return coins, nil
}

// MsgSend represents a send message in the banker.
type MsgSend struct {
	ToAddress crypto.Address // Send to address
	Send      string         // Send amount
}

// getType returns the type of the MsgSend.
func (msg MsgSend) GetType() string {
	return MSG_SEND
}

// validateMsg validates the MsgSend.
func (msg MsgSend) IsValid() error {
	if msg.ToAddress.IsZero() {
		return ErrInvalidToAddress
	}
	if _, err := std.ParseCoins(msg.Send); err != nil {
		return ErrInvalidAmount
	}
	return nil
}

// getCoins retrieves the coins involved in the MsgSend.
func (msg MsgSend) GetCoins() (std.Coins, error) {
	coins, err := std.ParseCoins(msg.Send)
	if err != nil {
		return nil, ErrInvalidAmount
	}
	return coins, nil
}

// MsgRun represents a run message in the VM.
type MsgRun struct {
	Package *std.MemPackage // Package to run
	Send    string          // Send amount
}

// getType returns the type of the MsgRun.
func (msg MsgRun) GetType() string {
	return MSG_RUN
}

// validateMsg validates the MsgRun.
func (msg MsgRun) IsValid() error {
	if msg.Package == nil || len(msg.Package.Files) == 0 {
		return ErrEmptyPackage
	}
	return nil
}

// getCoins retrieves the coins involved in the MsgRun.
func (msg MsgRun) GetCoins() (std.Coins, error) {
	coins, err := std.ParseCoins(msg.Send)
	if err != nil {
		return nil, ErrInvalidAmount
	}
	return coins, nil
}

// MsgAddPackage represents an add package message in the VM.
type MsgAddPackage struct {
	Package *std.MemPackage // Package to add
	Deposit string          // Coin deposit
}

// getType returns the type of the MsgAddPackage.
func (msg MsgAddPackage) GetType() string {
	return MSG_ADD_PKG
}

// validateMsg validates the MsgAddPackage.
func (msg MsgAddPackage) IsValid() error {
	if msg.Package == nil || len(msg.Package.Files) == 0 {
		return ErrEmptyPackage
	}
	return nil
}

// getCoins retrieves the coins involved in the MsgAddPackage.
func (msg MsgAddPackage) GetCoins() (std.Coins, error) {
	coins, err := std.ParseCoins(msg.Deposit)
	if err != nil {
		return nil, ErrInvalidAmount
	}
	return coins, nil
}

package gnoclient

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

var (
	ErrInvalidGasWanted  = errors.New("invalid gas wanted")
	ErrInvalidGasFee     = errors.New("invalid gas fee")
	ErrMissingSigner     = errors.New("missing Signer")
	ErrMissingRPCClient  = errors.New("missing RPCClient")
	ErrNoMessages        = errors.New("no messages provided")
	ErrMixedMessageTypes = errors.New("mixed message types not allowed")

	ErrInvalidSponsorAddress = errors.New("invalid sponsor address")
	ErrInvalidSponsorTx      = errors.New("invalid sponsor tx")
)

// BaseTxCfg defines the base transaction configuration shared by all message types.
type BaseTxCfg struct {
	GasFee         string // Gas fee
	GasWanted      int64  // Gas wanted
	AccountNumber  uint64 // Account number
	SequenceNumber uint64 // Sequence number
	Memo           string // Memo
}

func (cfg BaseTxCfg) validate() error {
	if cfg.GasWanted <= 0 {
		return ErrInvalidGasWanted
	}
	if cfg.GasFee == "" {
		return ErrInvalidGasFee
	}
	return nil
}

// SponsorTxCfg represents the configuration for a sponsor transaction.
type SponsorTxCfg struct {
	BaseTxCfg
	SponsorAddress crypto.Address
}

// IsValid validates the base transaction configuration.
func (cfg SponsorTxCfg) validate() error {
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

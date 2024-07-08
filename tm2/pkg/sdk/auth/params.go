package auth

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

type AuthParamsContextKey struct{}

// Default parameter values
const (
	DefaultMaxMemoBytes           int64 = 65536
	DefaultTxSigLimit             int64 = 7
	DefaultTxSizeCostPerByte      int64 = 10
	DefaultSigVerifyCostED25519   int64 = 590
	DefaultSigVerifyCostSecp256k1 int64 = 1000
)

// Params defines the parameters for the auth module.
type Params struct {
	MaxMemoBytes           int64 `json:"max_memo_bytes" yaml:"max_memo_bytes"`
	TxSigLimit             int64 `json:"tx_sig_limit" yaml:"tx_sig_limit"`
	TxSizeCostPerByte      int64 `json:"tx_size_cost_per_byte" yaml:"tx_size_cost_per_byte"`
	SigVerifyCostED25519   int64 `json:"sig_verify_cost_ed25519" yaml:"sig_verify_cost_ed25519"`
	SigVerifyCostSecp256k1 int64 `json:"sig_verify_cost_secp256k1" yaml:"sig_verify_cost_secp256k1"`
}

// NewParams creates a new Params object
func NewParams(maxMemoBytes, txSigLimit, txSizeCostPerByte,
	sigVerifyCostED25519, sigVerifyCostSecp256k1 int64,
) Params {
	return Params{
		MaxMemoBytes:           maxMemoBytes,
		TxSigLimit:             txSigLimit,
		TxSizeCostPerByte:      txSizeCostPerByte,
		SigVerifyCostED25519:   sigVerifyCostED25519,
		SigVerifyCostSecp256k1: sigVerifyCostSecp256k1,
	}
}

// Equals returns a boolean determining if two Params types are identical.
func (p Params) Equals(p2 Params) bool {
	return amino.DeepEqual(p, p2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxMemoBytes:           DefaultMaxMemoBytes,
		TxSigLimit:             DefaultTxSigLimit,
		TxSizeCostPerByte:      DefaultTxSizeCostPerByte,
		SigVerifyCostED25519:   DefaultSigVerifyCostED25519,
		SigVerifyCostSecp256k1: DefaultSigVerifyCostSecp256k1,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("MaxMemoBytes: %d\n", p.MaxMemoBytes))
	sb.WriteString(fmt.Sprintf("TxSigLimit: %d\n", p.TxSigLimit))
	sb.WriteString(fmt.Sprintf("TxSizeCostPerByte: %d\n", p.TxSizeCostPerByte))
	sb.WriteString(fmt.Sprintf("SigVerifyCostED25519: %d\n", p.SigVerifyCostED25519))
	sb.WriteString(fmt.Sprintf("SigVerifyCostSecp256k1: %d\n", p.SigVerifyCostSecp256k1))
	return sb.String()
}

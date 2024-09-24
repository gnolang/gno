package auth

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type AuthParamsContextKey struct{}

// Default parameter values
const (
	DefaultMaxMemoBytes              int64 = 65536
	DefaultTxSigLimit                int64 = 7
	DefaultTxSizeCostPerByte         int64 = 10
	DefaultSigVerifyCostED25519      int64 = 590
	DefaultSigVerifyCostSecp256k1    int64 = 1000
	DefaultGasPricesChangeCompressor int64 = 10
	DefaultTargetGasRatio            int64 = 70 //  70% of the MaxGas in a block
)

// Params defines the parameters for the auth module.
type Params struct {
	MaxMemoBytes              int64        `json:"max_memo_bytes" yaml:"max_memo_bytes"`
	TxSigLimit                int64        `json:"tx_sig_limit" yaml:"tx_sig_limit"`
	TxSizeCostPerByte         int64        `json:"tx_size_cost_per_byte" yaml:"tx_size_cost_per_byte"`
	SigVerifyCostED25519      int64        `json:"sig_verify_cost_ed25519" yaml:"sig_verify_cost_ed25519"`
	SigVerifyCostSecp256k1    int64        `json:"sig_verify_cost_secp256k1" yaml:"sig_verify_cost_secp256k1"`
	GasPricesChangeCompressor int64        `json:"gas_price_change_compressor" yaml:"gas_price_change_compressor"`
	TargetGasRatio            int64        `json:"target_gas_ratio" yaml:"target_gas_ratio"`
	InitialGasPrice           std.GasPrice `json:"initial_gasprice"`
}

// NewParams creates a new Params object
func NewParams(maxMemoBytes, txSigLimit, txSizeCostPerByte,
	sigVerifyCostED25519, sigVerifyCostSecp256k1, gasPricesChangeCompressor, targetGasRatio int64,
) Params {
	return Params{
		MaxMemoBytes:              maxMemoBytes,
		TxSigLimit:                txSigLimit,
		TxSizeCostPerByte:         txSizeCostPerByte,
		SigVerifyCostED25519:      sigVerifyCostED25519,
		SigVerifyCostSecp256k1:    sigVerifyCostSecp256k1,
		GasPricesChangeCompressor: gasPricesChangeCompressor,
		TargetGasRatio:            targetGasRatio,
	}
}

// Equals returns a boolean determining if two Params types are identical.
func (p Params) Equals(p2 Params) bool {
	return amino.DeepEqual(p, p2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxMemoBytes:              DefaultMaxMemoBytes,
		TxSigLimit:                DefaultTxSigLimit,
		TxSizeCostPerByte:         DefaultTxSizeCostPerByte,
		SigVerifyCostED25519:      DefaultSigVerifyCostED25519,
		SigVerifyCostSecp256k1:    DefaultSigVerifyCostSecp256k1,
		GasPricesChangeCompressor: DefaultGasPricesChangeCompressor,
		TargetGasRatio:            DefaultTargetGasRatio,
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
	sb.WriteString(fmt.Sprintf("GasPricesChangeCompressor: %d\n", p.GasPricesChangeCompressor))
	sb.WriteString(fmt.Sprintf("TargetGasRatio: %d\n", p.TargetGasRatio))
	return sb.String()
}

func (p Params) Validate() error {
	if p.TxSigLimit == 0 {
		return fmt.Errorf("invalid tx signature limit: %d", p.TxSigLimit)
	}
	if p.SigVerifyCostED25519 == 0 {
		return fmt.Errorf("invalid ED25519 signature verification cost: %d", p.SigVerifyCostED25519)
	}
	if p.SigVerifyCostSecp256k1 == 0 {
		return fmt.Errorf("invalid SECK256k1 signature verification cost: %d", p.SigVerifyCostSecp256k1)
	}
	if p.TxSizeCostPerByte == 0 {
		return fmt.Errorf("invalid tx size cost per byte: %d", p.TxSizeCostPerByte)
	}
	if p.GasPricesChangeCompressor <= 0 {
		return fmt.Errorf("invalid gas prices change compressor: %d, it should be larger or equal to 1", p.GasPricesChangeCompressor)
	}
	if p.TargetGasRatio < 0 || p.TargetGasRatio > 100 {
		return fmt.Errorf("invalid target block gas ratio: %d, it should be between 0 and 100, 0 is unlimited", p.TargetGasRatio)
	}
	return nil
}

func (ak AccountKeeper) SetParams(ctx sdk.Context, params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	ak.paramk.SetParam(ctx, ModuleName, params)
	return nil
}

func (ak AccountKeeper) GetParams(ctx sdk.Context) Params {
	params := &Params{}

	ok, err := ak.paramk.GetParam(ctx, ModuleName, params)

	if !ok {
		panic("params key " + ModuleName + " does not exist")
	}
	if err != nil {
		panic(err.Error())
	}
	return *params
}

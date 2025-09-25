package auth

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type AuthParamsContextKey struct{}

// Default parameter values
const (
	DefaultMaxMemoBytes           int64 = 65536
	DefaultTxSigLimit             int64 = 7
	DefaultTxSizeCostPerByte      int64 = 10
	DefaultSigVerifyCostED25519   int64 = 590
	DefaultSigVerifyCostSecp256k1 int64 = 1000

	DefaultGasPricesChangeCompressor int64 = 10
	DefaultTargetGasRatio            int64 = 70 //  70% of the MaxGas in a block

	DefaultFeeCollectorName string = "fee_collector"
)

// Params defines the parameters for the auth module.
type Params struct {
	MaxMemoBytes              int64            `json:"max_memo_bytes" yaml:"max_memo_bytes"`
	TxSigLimit                int64            `json:"tx_sig_limit" yaml:"tx_sig_limit"`
	TxSizeCostPerByte         int64            `json:"tx_size_cost_per_byte" yaml:"tx_size_cost_per_byte"`
	SigVerifyCostED25519      int64            `json:"sig_verify_cost_ed25519" yaml:"sig_verify_cost_ed25519"`
	SigVerifyCostSecp256k1    int64            `json:"sig_verify_cost_secp256k1" yaml:"sig_verify_cost_secp256k1"`
	GasPricesChangeCompressor int64            `json:"gas_price_change_compressor" yaml:"gas_price_change_compressor"`
	TargetGasRatio            int64            `json:"target_gas_ratio" yaml:"target_gas_ratio"`
	InitialGasPrice           std.GasPrice     `json:"initial_gasprice"`
	UnrestrictedAddrs         []crypto.Address `json:"unrestricted_addrs" yaml:"unrestricted_addrs"`
	FeeCollector              crypto.Address   `json:"fee_collector" yaml:"fee_collector"`
}

// NewParams creates a new Params object
func NewParams(maxMemoBytes, txSigLimit, txSizeCostPerByte,
	sigVerifyCostED25519, sigVerifyCostSecp256k1, gasPricesChangeCompressor, targetGasRatio int64,
	feeCollector crypto.Address,
) Params {
	return Params{
		MaxMemoBytes:              maxMemoBytes,
		TxSigLimit:                txSigLimit,
		TxSizeCostPerByte:         txSizeCostPerByte,
		SigVerifyCostED25519:      sigVerifyCostED25519,
		SigVerifyCostSecp256k1:    sigVerifyCostSecp256k1,
		GasPricesChangeCompressor: gasPricesChangeCompressor,
		TargetGasRatio:            targetGasRatio,
		FeeCollector:              feeCollector,
	}
}

// Equals returns a boolean determining if two Params types are identical.
func (p Params) Equals(p2 Params) bool {
	return amino.DeepEqual(p, p2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultMaxMemoBytes,
		DefaultTxSigLimit,
		DefaultTxSizeCostPerByte,
		DefaultSigVerifyCostED25519,
		DefaultSigVerifyCostSecp256k1,
		DefaultGasPricesChangeCompressor,
		DefaultTargetGasRatio,
		crypto.AddressFromPreimage([]byte(DefaultFeeCollectorName)),
	)
}

// String implements the stringer interface.
func (p Params) String() string {
	var builder strings.Builder
	sb := &builder // Pointer for use with fmt.Fprintf
	sb.WriteString("Params: \n")
	fmt.Fprintf(sb, "MaxMemoBytes: %d\n", p.MaxMemoBytes)
	fmt.Fprintf(sb, "TxSigLimit: %d\n", p.TxSigLimit)
	fmt.Fprintf(sb, "TxSizeCostPerByte: %d\n", p.TxSizeCostPerByte)
	fmt.Fprintf(sb, "SigVerifyCostED25519: %d\n", p.SigVerifyCostED25519)
	fmt.Fprintf(sb, "SigVerifyCostSecp256k1: %d\n", p.SigVerifyCostSecp256k1)
	fmt.Fprintf(sb, "GasPricesChangeCompressor: %d\n", p.GasPricesChangeCompressor)
	fmt.Fprintf(sb, "TargetGasRatio: %d\n", p.TargetGasRatio)
	fmt.Fprintf(sb, "FeeCollector: %s\n", p.FeeCollector.String())
	return sb.String()
}

func (p Params) Validate() error {
	if p.MaxMemoBytes <= 0 {
		return fmt.Errorf("invalid max memo bytes: %d", p.MaxMemoBytes)
	}
	if p.TxSigLimit <= 0 {
		return fmt.Errorf("invalid tx signature limit: %d", p.TxSigLimit)
	}
	if p.SigVerifyCostED25519 <= 0 {
		return fmt.Errorf("invalid ED25519 signature verification cost: %d", p.SigVerifyCostED25519)
	}
	if p.SigVerifyCostSecp256k1 <= 0 {
		return fmt.Errorf("invalid SECK256k1 signature verification cost: %d", p.SigVerifyCostSecp256k1)
	}
	if p.TxSizeCostPerByte <= 0 {
		return fmt.Errorf("invalid tx size cost per byte: %d", p.TxSizeCostPerByte)
	}
	if p.GasPricesChangeCompressor <= 0 {
		return fmt.Errorf("invalid gas prices change compressor: %d, it should be larger or equal to 1", p.GasPricesChangeCompressor)
	}
	if p.TargetGasRatio < 0 || p.TargetGasRatio > 100 {
		return fmt.Errorf("invalid target block gas ratio: %d, it should be between 0 and 100, 0 is unlimited", p.TargetGasRatio)
	}
	if p.FeeCollector.IsZero() {
		return fmt.Errorf("invalid fee collector, cannot be empty")
	}
	return nil
}

const (
	// feeCollectorPath the params path for the fee collector account address
	feeCollectorPath = "p:fee_collector"
)

func (ak AccountKeeper) FeeCollectorAddress(ctx sdk.Context) crypto.Address {
	feeCollector := ak.GetParams(ctx).FeeCollector
	if feeCollector.IsZero() {
		panic("empty `fee_collector` param value")
	}
	return feeCollector
}

func (ak AccountKeeper) SetFeesCollectorAddress(ctx sdk.Context, addr crypto.Address) {
	ak.prmk.SetString(ctx, feeCollectorPath, addr.Bech32().String())
}

func (ak AccountKeeper) SetParams(ctx sdk.Context, params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	ak.prmk.SetStruct(ctx, "p", params)
	return nil
}

func (ak AccountKeeper) GetParams(ctx sdk.Context) Params {
	params := Params{}
	ak.prmk.GetStruct(ctx, "p", &params)
	return params
}

// WillSetParam defines what needs to be done when the parameter is set.
func (ak AccountKeeper) WillSetParam(ctx sdk.Context, key string, value any) {
	logger := ak.Logger(ctx)
	switch key {
	case "p:unrestricted_addrs":
		addrs, ok := value.([]string)
		if !ok {
			return
		}
		ak.applyUnrestrictedAddrsChange(ctx, addrs)
	default:
		// No-op for unrecognized keys
		logger.Error("No-op for unrecognized keys", "key", key)
	}
}

func (ak AccountKeeper) applyUnrestrictedAddrsChange(ctx sdk.Context, newAddrs []string) {
	params, ok := ctx.Value(AuthParamsContextKey{}).(Params)
	if !ok {
		panic("missing or invalid AuthParams in context")
	}
	// Build sets once.
	oldSet := make(map[string]struct{}, len(params.UnrestrictedAddrs))
	for _, addr := range params.UnrestrictedAddrs {
		oldSet[addr.String()] = struct{}{}
	}
	newSet := make(map[string]struct{}, len(newAddrs))
	for _, s := range newAddrs {
		newSet[s] = struct{}{}
	}
	// addition
	for s := range newSet {
		if _, ok := oldSet[s]; ok {
			continue // in both, no change
		}
		addr, err := crypto.AddressFromString(s)
		if err != nil {
			panic(fmt.Sprintf("invalid address: %v", err))
		}
		acc := ak.GetAccount(ctx, addr)
		uacc, ok := acc.(std.AccountUnrestricter)
		if !ok {
			continue
		}
		uacc.SetTokenLockWhitelisted(true)
		ak.SetAccount(ctx, acc)
	}
	// removal
	for s := range oldSet {
		if _, ok := newSet[s]; ok {
			continue // in both, no change
		}
		addr, err := crypto.AddressFromString(s)
		if err != nil {
			panic(fmt.Sprintf("invalid address: %v", err))
		}
		acc := ak.GetAccount(ctx, addr)
		uacc, ok := acc.(std.AccountUnrestricter)
		if !ok {
			continue
		}
		uacc.SetTokenLockWhitelisted(false)
		ak.SetAccount(ctx, acc)
	}
}

package auth

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name         string
		params       Params
		expectsError bool
	}{
		{
			name: "Valid Params",
			params: Params{
				MaxMemoBytes:              256,
				TxSigLimit:                7,
				TxSizeCostPerByte:         1,
				SigVerifyCostED25519:      100,
				SigVerifyCostSecp256k1:    200,
				GasPricesChangeCompressor: 1,
				TargetGasRatio:            50,
				FeeCollector:              crypto.AddressFromPreimage([]byte("test_collector")),
			},
			expectsError: false,
		},
		{
			// TxSigLimit counts leaves; multisig.MaxTotalKeys counts the whole
			// key and is a compile-time constant this parameter cannot raise.
			// A limit past the leaves that constant admits would let a chain
			// permit keys its own signature check rejects.
			name: "Invalid TxSigLimit above what multisig.MaxTotalKeys admits",
			params: Params{
				MaxMemoBytes:              256,
				TxSigLimit:                int64((multisig.MaxTotalKeys+1)/2) + 1,
				TxSizeCostPerByte:         1,
				SigVerifyCostED25519:      100,
				SigVerifyCostSecp256k1:    200,
				GasPricesChangeCompressor: 1,
				TargetGasRatio:            50,
				FeeCollector:              crypto.AddressFromPreimage([]byte("test_collector")),
			},
			expectsError: true,
		},
		{
			name: "Invalid TxSigLimit",
			params: Params{
				TxSigLimit: 0,
			},
			expectsError: true,
		},
		{
			name: "Invalid SigVerifyCostED25519",
			params: Params{
				SigVerifyCostED25519: 0,
			},
			expectsError: true,
		},
		{
			name: "Invalid GasPricesChangeCompressor",
			params: Params{
				GasPricesChangeCompressor: 0,
			},
			expectsError: true,
		},
		{
			name: "Invalid TargetGasRatio",
			params: Params{
				TargetGasRatio: 150,
			},
			expectsError: true,
		},
		{
			name: "Invalid SigVerifyCostSecp256k1",
			params: Params{
				SigVerifyCostSecp256k1: 0,
			},
			expectsError: true,
		},
		{
			name: "Invalid TxSizeCostPerByte",
			params: Params{
				MaxMemoBytes:           256,
				TxSigLimit:             10,
				SigVerifyCostED25519:   100,
				SigVerifyCostSecp256k1: 200,
				TxSizeCostPerByte:      -1,
			},
			expectsError: true,
		},
		{
			name: "Invalid FeeCollector empty",
			params: Params{
				MaxMemoBytes:              256,
				TxSigLimit:                10,
				TxSizeCostPerByte:         1,
				SigVerifyCostED25519:      100,
				SigVerifyCostSecp256k1:    200,
				GasPricesChangeCompressor: 1,
				TargetGasRatio:            50,
			},
			expectsError: true,
		},
		{
			name: "Invalid InitialGasPrice negative gas",
			params: Params{
				MaxMemoBytes:              256,
				TxSigLimit:                10,
				TxSizeCostPerByte:         1,
				SigVerifyCostED25519:      100,
				SigVerifyCostSecp256k1:    200,
				GasPricesChangeCompressor: 1,
				TargetGasRatio:            50,
				FeeCollector:              crypto.AddressFromPreimage([]byte("test_collector")),
				InitialGasPrice:           std.GasPrice{Gas: -1, Price: std.Coin{Denom: "ugnot", Amount: 1}},
			},
			expectsError: true,
		},
		{
			name: "Invalid InitialGasPrice negative price amount",
			params: Params{
				MaxMemoBytes:              256,
				TxSigLimit:                10,
				TxSizeCostPerByte:         1,
				SigVerifyCostED25519:      100,
				SigVerifyCostSecp256k1:    200,
				GasPricesChangeCompressor: 1,
				TargetGasRatio:            50,
				FeeCollector:              crypto.AddressFromPreimage([]byte("test_collector")),
				InitialGasPrice:           std.GasPrice{Gas: 1, Price: std.Coin{Denom: "ugnot", Amount: -5}},
			},
			expectsError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectsError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewParams(t *testing.T) {
	// Define expected values for each parameter
	maxMemoBytes := int64(256)
	txSigLimit := int64(10)
	txSizeCostPerByte := int64(5)
	sigVerifyCostED25519 := int64(100)
	sigVerifyCostSecp256k1 := int64(200)
	gasPricesChangeCompressor := int64(50)
	targetGasRatio := int64(75)
	feeCollector := crypto.AddressFromPreimage([]byte("test_collector"))

	// Call NewParams with the values
	params := NewParams(
		maxMemoBytes,
		txSigLimit,
		txSizeCostPerByte,
		sigVerifyCostED25519,
		sigVerifyCostSecp256k1,
		gasPricesChangeCompressor,
		targetGasRatio,
		feeCollector,
	)

	// Create an expected Params struct with the same values
	expectedParams := Params{
		MaxMemoBytes:              maxMemoBytes,
		TxSigLimit:                txSigLimit,
		TxSizeCostPerByte:         txSizeCostPerByte,
		SigVerifyCostED25519:      sigVerifyCostED25519,
		SigVerifyCostSecp256k1:    sigVerifyCostSecp256k1,
		GasPricesChangeCompressor: gasPricesChangeCompressor,
		TargetGasRatio:            targetGasRatio,
		FeeCollector:              feeCollector,
	}

	// Check if the returned params struct matches the expected struct
	if !reflect.DeepEqual(params, expectedParams) {
		t.Errorf("NewParams() = %+v, want %+v", params, expectedParams)
	}
}

func TestInitialGasPriceParamRoundTrip(t *testing.T) {
	env := setupTestEnv()
	want := std.GasPrice{Gas: 1, Price: std.Coin{Denom: "ugnot", Amount: 1}}

	env.acck.prmk.SetString(env.ctx, "p:initial_gasprice", "1ugnot/1gas")
	require.Equal(t, want, env.acck.GetParams(env.ctx).InitialGasPrice)

	params := DefaultParams()
	params.InitialGasPrice = want
	require.NoError(t, env.acck.SetParams(env.ctx, params))
	require.Equal(t, params, env.acck.GetParams(env.ctx))
}

func TestWillSetParam(t *testing.T) {
	env := setupTestEnv()

	tests := []struct {
		name        string
		key         string
		value       any
		shouldPanic bool
	}{
		// unrestricted_addrs
		{
			name:        "valid unrestricted_addrs",
			key:         "p:unrestricted_addrs",
			value:       []string{},
			shouldPanic: false,
		},
		{
			name:        "wrong type for unrestricted_addrs",
			key:         "p:unrestricted_addrs",
			value:       "not_a_slice",
			shouldPanic: true,
		},
		// fee_collector
		{
			name:        "valid fee_collector",
			key:         "p:fee_collector",
			value:       "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			shouldPanic: false,
		},
		{
			name:        "wrong type for fee_collector",
			key:         "p:fee_collector",
			value:       int64(123),
			shouldPanic: true,
		},
		{
			name:        "invalid fee_collector address",
			key:         "p:fee_collector",
			value:       "invalid_address",
			shouldPanic: true,
		},
		// int64 params
		{
			name:        "valid max_memo_bytes",
			key:         "p:max_memo_bytes",
			value:       int64(1024),
			shouldPanic: false,
		},
		{
			name:        "wrong type for max_memo_bytes",
			key:         "p:max_memo_bytes",
			value:       "not_int64",
			shouldPanic: true,
		},
		{
			name:        "invalid max_memo_bytes value",
			key:         "p:max_memo_bytes",
			value:       int64(0),
			shouldPanic: true,
		},
		{
			name:        "valid tx_sig_limit",
			key:         "p:tx_sig_limit",
			value:       int64((multisig.MaxTotalKeys + 1) / 2),
			shouldPanic: false,
		},
		{
			// Governance can raise this parameter; it cannot raise
			// multisig.MaxTotalKeys, so a value past what that constant admits
			// has to be refused here rather than desynchronising the two.
			name:        "tx_sig_limit above what multisig.MaxTotalKeys admits panics",
			key:         "p:tx_sig_limit",
			value:       int64((multisig.MaxTotalKeys+1)/2) + 1,
			shouldPanic: true,
		},
		{
			name:        "wrong type for tx_sig_limit",
			key:         "p:tx_sig_limit",
			value:       "not_int64",
			shouldPanic: true,
		},
		// tx_size_cost_per_byte
		{
			name:        "valid tx_size_cost_per_byte",
			key:         "p:tx_size_cost_per_byte",
			value:       int64(20),
			shouldPanic: false,
		},
		{
			name:        "wrong type for tx_size_cost_per_byte",
			key:         "p:tx_size_cost_per_byte",
			value:       "not_int64",
			shouldPanic: true,
		},
		{
			name:        "invalid tx_size_cost_per_byte value",
			key:         "p:tx_size_cost_per_byte",
			value:       int64(0),
			shouldPanic: true,
		},
		// sig_verify_cost_ed25519
		{
			name:        "valid sig_verify_cost_ed25519",
			key:         "p:sig_verify_cost_ed25519",
			value:       int64(500),
			shouldPanic: false,
		},
		{
			name:        "wrong type for sig_verify_cost_ed25519",
			key:         "p:sig_verify_cost_ed25519",
			value:       "not_int64",
			shouldPanic: true,
		},
		{
			name:        "invalid sig_verify_cost_ed25519 value",
			key:         "p:sig_verify_cost_ed25519",
			value:       int64(-1),
			shouldPanic: true,
		},
		// sig_verify_cost_secp256k1
		{
			name:        "valid sig_verify_cost_secp256k1",
			key:         "p:sig_verify_cost_secp256k1",
			value:       int64(1000),
			shouldPanic: false,
		},
		{
			name:        "wrong type for sig_verify_cost_secp256k1",
			key:         "p:sig_verify_cost_secp256k1",
			value:       "not_int64",
			shouldPanic: true,
		},
		{
			name:        "invalid sig_verify_cost_secp256k1 value",
			key:         "p:sig_verify_cost_secp256k1",
			value:       int64(0),
			shouldPanic: true,
		},
		// gas_price_change_compressor
		{
			name:        "valid gas_price_change_compressor",
			key:         "p:gas_price_change_compressor",
			value:       int64(15),
			shouldPanic: false,
		},
		{
			name:        "wrong type for gas_price_change_compressor",
			key:         "p:gas_price_change_compressor",
			value:       "not_int64",
			shouldPanic: true,
		},
		{
			name:        "invalid gas_price_change_compressor value",
			key:         "p:gas_price_change_compressor",
			value:       int64(0),
			shouldPanic: true,
		},
		{
			name:        "valid target_gas_ratio",
			key:         "p:target_gas_ratio",
			value:       int64(70),
			shouldPanic: false,
		},
		{
			name:        "invalid target_gas_ratio value",
			key:         "p:target_gas_ratio",
			value:       int64(150),
			shouldPanic: true,
		},
		// initial_gasprice
		{
			name:        "valid initial_gasprice",
			key:         "p:initial_gasprice",
			value:       "1ugnot/1gas",
			shouldPanic: false,
		},
		{
			name:        "wrong type for initial_gasprice",
			key:         "p:initial_gasprice",
			value:       int64(123),
			shouldPanic: true,
		},
		{
			name:        "invalid initial_gasprice format",
			key:         "p:initial_gasprice",
			value:       "invalid",
			shouldPanic: true,
		},
		// unknown key
		{
			name:        "unknown param key panics",
			key:         "p:nonexistent",
			value:       "foo",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				require.Panics(t, func() {
					env.acck.WillSetParam(env.ctx, tt.key, tt.value)
				})
			} else {
				require.NotPanics(t, func() {
					env.acck.WillSetParam(env.ctx, tt.key, tt.value)
				})
			}
		})
	}
}

// TestWillSetParamExhaustive ensures every Params field has a WillSetParam case.
func TestWillSetParamExhaustive(t *testing.T) {
	env := setupTestEnv()

	call := func(param string) (pnc any) {
		defer func() {
			pnc = recover()
		}()
		env.acck.WillSetParam(env.ctx, param, "")
		return nil
	}

	// baseline: ensure a non-existent key has the expected error.
	const format = "unknown auth param key: %q"
	assert.Equal(t, fmt.Sprintf(format, "doesnotexist"), call("doesnotexist"))

	typ := reflect.TypeFor[Params]()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag, _, _ := strings.Cut(field.Tag.Get("json"), ",")

		t.Run(jsonTag, func(t *testing.T) {
			assert.NotEqual(t, fmt.Sprintf(format, "p:"+jsonTag), call("p:"+jsonTag))
		})
	}
}

func TestParamsString(t *testing.T) {
	cases := []struct {
		name   string
		params Params
		want   string
	}{
		{"blank params", Params{}, "Params: \nMaxMemoBytes: 0\nTxSigLimit: 0\nTxSizeCostPerByte: 0\nSigVerifyCostED25519: 0\nSigVerifyCostSecp256k1: 0\nGasPricesChangeCompressor: 0\nTargetGasRatio: 0\nFeeCollector: g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe\n"},
		{"some values", Params{
			MaxMemoBytes:      1_000_000,
			TxSizeCostPerByte: 8192,
		}, "Params: \nMaxMemoBytes: 1000000\nTxSigLimit: 0\nTxSizeCostPerByte: 8192\nSigVerifyCostED25519: 0\nSigVerifyCostSecp256k1: 0\nGasPricesChangeCompressor: 0\nTargetGasRatio: 0\nFeeCollector: g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe\n"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.params.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInitialGasPriceValidation(t *testing.T) {
	for _, tc := range []struct {
		value string
		// Whether Params.Validate rejects the value on its own. A foreign
		// denomination does not fail Validate -- tm2 does not know which denom
		// the hosted chain collects fees in, so it only checks the shape -- but
		// governance still cannot introduce one, because WillSetParam pins the
		// denom to the one already configured.
		invalidAsParams bool
	}{
		{"9223372036854775807ugnot/9223372036854775807gas", true},
		{"1000000foo/1gas", false},
		{"1000000000001ugnot/1gas", true},
		{"1ugnot/1000000000001gas", true},
		// A zero price amount would pin the block gas price at zero for good:
		// calcBlockGasPrice returns early on a zero amount, so no later
		// proposal can restore a working fee floor.
		{"0ugnot/1gas", true},
	} {
		value := tc.value
		t.Run(value, func(t *testing.T) {
			env := setupTestEnv()
			env.acck.prmk.SetString(env.ctx, "p:initial_gasprice", "1ugnot/1000gas")
			before := env.acck.GetParams(env.ctx)
			require.Panics(t, func() { env.acck.prmk.SetString(env.ctx, "p:initial_gasprice", value) })
			require.Equal(t, before, env.acck.GetParams(env.ctx))
			bad := before
			var err error
			bad.InitialGasPrice, err = std.ParseGasPrice(value)
			require.NoError(t, err)
			if tc.invalidAsParams {
				require.Error(t, env.acck.SetParams(env.ctx, bad))
			}

			env.gk.SetGasPrice(env.ctx, before.InitialGasPrice)
			for _, used := range []int64{0, 10_000_000, 0} {
				ctx := env.ctx.WithValue(AuthParamsContextKey{}, env.acck.GetParams(env.ctx)).WithBlockGasMeter(store.NewGasMeter(10_000_000))
				ctx.BlockGasMeter().ConsumeGas(used, "block")
				require.NotPanics(t, func() { EndBlocker(ctx, env.gk) })
				ctx = ctx.WithValue(GasPriceContextKey{}, env.gk.LastGasPrice(ctx))
				res := EnsureSufficientMempoolFees(ctx, std.NewFee(1000, std.NewCoin("ugnot", 1000)))
				require.True(t, res.IsOK(), res.Log)
			}
		})
	}
	for _, value := range []string{"1ugnot/1gas", "1ugnot/1000gas", "1000000000000ugnot/1000000000000gas"} {
		env := setupTestEnv()
		require.NotPanics(t, func() { env.acck.prmk.SetString(env.ctx, "p:initial_gasprice", value) })
		want, err := std.ParseGasPrice(value)
		require.NoError(t, err)
		require.Equal(t, want, env.acck.GetParams(env.ctx).InitialGasPrice)
	}
}

// TestInitialGasPriceDenomIsPinned checks that governance can move the gas
// price but not the denomination it is quoted in. Switching it is unrecoverable:
// std.GasPrice.IsGTE refuses to compare prices across denoms, so every
// transaction fails EnsureSufficientMempoolFees, including the proposal that
// would switch it back.
func TestInitialGasPriceDenomIsPinned(t *testing.T) {
	// Whatever genesis chose is the denom governance is held to -- tm2 does not
	// assume it is ugnot.
	for _, genesisDenom := range []string{"ugnot", "atom"} {
		t.Run(genesisDenom, func(t *testing.T) {
			env := setupTestEnv()
			p := DefaultParams()
			p.InitialGasPrice = std.GasPrice{Gas: 1000, Price: std.Coin{Denom: genesisDenom, Amount: 1}}
			require.NoError(t, env.acck.SetParams(env.ctx, p))

			// Same denom, different price: allowed.
			require.NotPanics(t, func() {
				env.acck.prmk.SetString(env.ctx, "p:initial_gasprice", "5"+genesisDenom+"/1000gas")
			})
			require.Equal(t, int64(5), env.acck.GetParams(env.ctx).InitialGasPrice.Price.Amount)

			// Different denom: refused, and nothing is written.
			before := env.acck.GetParams(env.ctx)
			require.Panics(t, func() {
				env.acck.prmk.SetString(env.ctx, "p:initial_gasprice", "1zzz/1000gas")
			})
			require.Equal(t, before, env.acck.GetParams(env.ctx))
		})
	}

	// An unset price has no denom to preserve, so first enabling pricing may
	// pick one.
	env := setupTestEnv()
	require.NoError(t, env.acck.SetParams(env.ctx, DefaultParams()))
	require.NotPanics(t, func() {
		env.acck.prmk.SetString(env.ctx, "p:initial_gasprice", "1ugnot/1000gas")
	})
}

// TestInitialGasPriceStoreRoundTripStaysValid pins the invariant that a gas
// price Validate accepts must still be accepted after a trip through the params
// store. The store is lossy for a zero amount -- it marshals a Coin through
// Coin.String(), which renders "" and drops the denom -- so a rule keyed on the
// denom can accept a write and reject the read of the same value. WillSetParam
// re-validates the whole struct it reads back, so that asymmetry panics every
// later write to any auth param, and ExportGenesis emits an unloadable genesis.
func TestInitialGasPriceStoreRoundTripStaysValid(t *testing.T) {
	for _, gp := range []std.GasPrice{
		{},
		{Gas: 0, Price: std.Coin{Denom: "ugnot", Amount: 0}},
		{Gas: 1000, Price: std.Coin{Denom: "ugnot", Amount: 1}},
		{Gas: 1, Price: std.Coin{Denom: "ugnot", Amount: MaxGasPriceComponent}},
	} {
		t.Run(gp.String(), func(t *testing.T) {
			env := setupTestEnv()
			p := DefaultParams()
			p.InitialGasPrice = gp
			require.NoError(t, env.acck.SetParams(env.ctx, p))

			readBack := env.acck.GetParams(env.ctx)
			require.NoError(t, readBack.Validate(),
				"a stored gas price must still validate when read back")
			require.NoError(t, env.acck.ExportGenesis(env.ctx).Params.Validate(),
				"an exported genesis must be loadable again")

			// The read-back value is what WillSetParam validates, so a write to
			// an unrelated auth param has to go through.
			require.NotPanics(t, func() {
				env.acck.prmk.SetInt64(env.ctx, "p:max_memo_bytes", 32768)
			})
		})
	}
}

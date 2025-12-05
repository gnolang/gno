package auth

import (
	"reflect"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
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
				TxSigLimit:                10,
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

func TestSetParams(t *testing.T) {
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

	// set params
	env := setupTestEnv()
	env.acck.SetParams(env.ctx, params)

	params2 := env.acck.GetParams(env.ctx)
	assert.Equal(t, params, params2, "params got not equal")
}

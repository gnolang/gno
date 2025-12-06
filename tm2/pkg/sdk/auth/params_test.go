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
	gasPricesChangeCompressor := int64(50)
	targetGasRatio := int64(75)
	feeCollector := crypto.AddressFromPreimage([]byte("test_collector"))

	// Call NewParams with the values
	params := NewParams(
		maxMemoBytes,
		txSigLimit,
		gasPricesChangeCompressor,
		targetGasRatio,
		feeCollector,
	)

	// Create an expected Params struct with the same values
	expectedParams := Params{
		MaxMemoBytes:              maxMemoBytes,
		TxSigLimit:                txSigLimit,
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
			MaxMemoBytes: 1_000_000,
		}, "Params: \nMaxMemoBytes: 1000000\nTxSigLimit: 0\nTxSizeCostPerByte: 8192\nSigVerifyCostED25519: 0\nSigVerifyCostSecp256k1: 0\nGasPricesChangeCompressor: 0\nTargetGasRatio: 0\nFeeCollector: g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe\n"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.params.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

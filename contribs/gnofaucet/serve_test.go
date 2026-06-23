package main

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeFaucet_CleanupShorterThanRateLimit(t *testing.T) {
	t.Parallel()

	cfg := &serveCfg{
		rateLimitInterval:     24 * time.Hour,
		rateLimitCleanTimeout: time.Hour,
	}

	err := serveFaucet(context.Background(), cfg, nil)

	assert.ErrorContains(t, err, "ratelimit-cleanup-timeout must be >= ratelimit-interval")
}

func TestServeFaucet_InvalidGasValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		gasFee      string
		gasWanted   int64
		expectedErr string
	}{
		{
			name:        "malformed gas fee",
			gasFee:      "invalid",
			gasWanted:   100000,
			expectedErr: "invalid gas fee",
		},
		{
			name:        "zero gas wanted",
			gasFee:      "1000000ugnot",
			gasWanted:   0,
			expectedErr: "gas wanted must be greater than zero",
		},
		{
			name:        "negative gas wanted",
			gasFee:      "1000000ugnot",
			gasWanted:   -100,
			expectedErr: "gas wanted must be greater than zero",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := &serveCfg{
				rateLimitInterval:     time.Hour,
				rateLimitCleanTimeout: 24 * time.Hour,
				gasFee:                testCase.gasFee,
				gasWanted:             testCase.gasWanted,
			}

			err := serveFaucet(context.Background(), cfg, nil)

			assert.ErrorContains(t, err, testCase.expectedErr)
		})
	}
}

func TestServeCfg_GasFlags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		args              []string
		expectedGasFee    string
		expectedGasWanted int64
	}{
		{
			name:              "defaults",
			args:              nil,
			expectedGasFee:    "1000000ugnot",
			expectedGasWanted: 100000,
		},
		{
			name: "explicit values",
			args: []string{
				"-gas-fee", "2000000ugnot",
				"-gas-wanted", "2000000",
			},
			expectedGasFee:    "2000000ugnot",
			expectedGasWanted: 2000000,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := &serveCfg{}
			fs := flag.NewFlagSet("serve", flag.ContinueOnError)
			cfg.RegisterFlags(fs)

			require.NoError(t, fs.Parse(testCase.args))

			assert.Equal(t, testCase.expectedGasFee, cfg.gasFee)
			assert.Equal(t, testCase.expectedGasWanted, cfg.gasWanted)
		})
	}
}

package client

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGasWanted(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		{
			name:     "valid gas amount",
			input:    "1000000",
			expected: 1000000,
			wantErr:  false,
		},
		{
			name:     "zero gas",
			input:    "0",
			expected: 0,
			wantErr:  false,
		},
		{
			name:    "invalid format",
			input:   "abc",
			wantErr: true,
		},
		{
			name:     "negative gas",
			input:    "-100",
			expected: -100,
			wantErr:  false, // ParseInt allows negative numbers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGasWanted(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMakeTxCfg_GasAutoFlag(t *testing.T) {
	tests := []struct {
		name        string
		gasInput    string
		expectedAuto bool
		expectedGas  int64
		wantErr     bool
	}{
		{
			name:        "auto gas",
			gasInput:    "auto",
			expectedAuto: true,
			expectedGas:  0,
			wantErr:     false,
		},
		{
			name:        "numeric gas",
			gasInput:    "1000000",
			expectedAuto: false,
			expectedGas:  1000000,
			wantErr:     false,
		},
		{
			name:        "empty string",
			gasInput:    "",
			expectedAuto: false,
			expectedGas:  0,
			wantErr:     false,
		},
		{
			name:     "invalid gas",
			gasInput: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &MakeTxCfg{}
			
			// Simulate the flag parsing function
			err := func(s string) error {
				if s == "auto" {
					cfg.GasAuto = true
					cfg.GasWanted = 0
				} else if s != "" {
					gasWanted, err := parseGasWanted(s)
					if err != nil {
						return err
					}
					cfg.GasWanted = gasWanted
					cfg.GasAuto = false
				}
				return nil
			}(tt.gasInput)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAuto, cfg.GasAuto)
				assert.Equal(t, tt.expectedGas, cfg.GasWanted)
			}
		})
	}
}

func TestEstimateGasAndFee_NoAutoMode(t *testing.T) {
	cfg := &MakeTxCfg{
		GasAuto: false,
	}
	
	tx := &std.Tx{}
	
	// Should return immediately without error when auto mode is disabled
	err := EstimateGasAndFee(cfg, tx)
	require.NoError(t, err)
}

func TestEstimateGasAndFee_MissingRemote(t *testing.T) {
	cfg := &MakeTxCfg{
		GasAuto: true,
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Remote: "", // Missing remote URL
			},
		},
	}
	
	tx := &std.Tx{}
	
	// Should return error when remote URL is missing
	err := EstimateGasAndFee(cfg, tx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing remote url")
}
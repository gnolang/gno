package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGasPriceGTE(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		gp          GasPrice
		gpB         GasPrice
		expectError bool
		errorMsg    string
		expected    bool // for non-error cases: whether gp.IsGTE(gpB) should return true or false
	}{
		// Error cases: Different denominations
		{
			name: "Different denominations error",
			gp: GasPrice{
				Gas: 100,
				Price: Coin{
					Denom:  "atom",
					Amount: 500,
				},
			},
			gpB: GasPrice{
				Gas: 100,
				Price: Coin{
					Denom:  "btc", // Different denomination
					Amount: 500,
				},
			},
			expectError: true,
			errorMsg:    "Gas price denominations should be equal;",
		},
		// Error cases: Zero Gas values
		{
			name: "Zero Gas in gp error",
			gp: GasPrice{
				Gas: 0, // Zero Gas in gp
				Price: Coin{
					Denom:  "atom",
					Amount: 500,
				},
			},
			gpB: GasPrice{
				Gas: 100,
				Price: Coin{
					Denom:  "atom",
					Amount: 500,
				},
			},
			expectError: true,
			errorMsg:    "GasPrice.Gas cannot be zero;",
		},
		{
			name: "Zero Gas in gpB error",
			gp: GasPrice{
				Gas: 100,
				Price: Coin{
					Denom:  "atom",
					Amount: 500,
				},
			},
			gpB: GasPrice{
				Gas: 0, // Zero Gas in gpB
				Price: Coin{
					Denom:  "atom",
					Amount: 500,
				},
			},
			expectError: true,
			errorMsg:    "GasPrice.Gas cannot be zero;",
		},
		// Valid cases: No errors, just compare gas prices
		{
			name: "Greater Gas Price",
			gp: GasPrice{
				Gas: 100,
				Price: Coin{
					Denom:  "atom",
					Amount: 600, // Greater price
				},
			},
			gpB: GasPrice{
				Gas: 100,
				Price: Coin{
					Denom:  "atom",
					Amount: 500,
				},
			},
			expectError: false,
			expected:    true,
		},
		{
			name: "Equal Gas Price",
			gp: GasPrice{
				Gas: 100,
				Price: Coin{
					Denom:  "atom",
					Amount: 500,
				},
			},
			gpB: GasPrice{
				Gas: 100,
				Price: Coin{
					Denom:  "atom",
					Amount: 500,
				},
			},
			expectError: false,
			expected:    true,
		},
		{
			name: "Lesser Gas Price",
			gp: GasPrice{
				Gas: 100,
				Price: Coin{
					Denom:  "atom",
					Amount: 400, // Lesser price
				},
			},
			gpB: GasPrice{
				Gas: 100,
				Price: Coin{
					Denom:  "atom",
					Amount: 500,
				},
			},
			expectError: false,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.gp.IsGTE(tt.gpB)
			if !tt.expectError {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got, "Expect that %v is less than %v", tt.gp, tt.gpB)
				if got != tt.expected {
					t.Errorf("Test %s failed: expected result %v, got %v", tt.name, tt.expected, got)
				}
			} else {
				require.Error(t, err)
				errorMsg := err.Error()
				assert.Contains(t, errorMsg, tt.errorMsg, "expected error message containing %q, but got %q", tt.errorMsg, errorMsg)
			}
		})
	}
}

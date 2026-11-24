package std

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGasPriceJSON(t *testing.T) {
	t.Parallel()

	want := GasPrice{Gas: 1, Price: Coin{Denom: "ugnot", Amount: 1}}

	var got GasPrice
	require.NoError(t, json.Unmarshal([]byte(`"1ugnot/1gas"`), &got))
	require.Equal(t, want, got)
	require.NoError(t, json.Unmarshal([]byte(`null`), &got))
	require.Equal(t, want, got)

	require.Error(t, json.Unmarshal([]byte(`"bogus"`), &got))

	bz, err := json.Marshal(want)
	require.NoError(t, err)
	require.Equal(t, `{"gas":1,"price":{"denom":"ugnot","amount":1}}`, string(bz))

	// Object form (from json.Marshal) round-trips via the alias branch.
	var fromObj GasPrice
	require.NoError(t, json.Unmarshal(bz, &fromObj))
	require.Equal(t, want, fromObj)
}

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

// IsGTE compares fee-per-gas by cross-multiplying, so a negative gas flips one
// side's sign and inverts the result: paying nothing reports as sufficient. The
// gas comes from the transaction, and Tx.ValidateBasic bounds gas_wanted only
// from above.
//
// Reaching it needs a negative to survive as far as the fee check, which it does
// not today -- SetGasMeter runs two lines later and NewGasMeter panics on a
// negative limit. That guard is in another package, so this one does not rely on
// it.
func TestGasPriceIsGTERefusesNegativeGas(t *testing.T) {
	t.Parallel()

	// A node minimum of 1ugnot per 1000 gas.
	nodeMin := GasPrice{Gas: 1000, Price: Coin{Denom: "ugnot", Amount: 1}}

	// The control: paying nothing for real gas is correctly insufficient.
	honest := GasPrice{Gas: 1000, Price: Coin{Denom: "ugnot", Amount: 0}}
	ok, err := honest.IsGTE(nodeMin)
	require.NoError(t, err)
	require.False(t, ok, "a zero fee for positive gas must be insufficient")

	// The same zero fee with the gas negated must not become sufficient.
	for _, gas := range []int64{-1, -1000, math.MinInt64} {
		free := GasPrice{Gas: gas, Price: Coin{Denom: "ugnot", Amount: 0}}
		ok, err := free.IsGTE(nodeMin)
		require.Error(t, err, "gas %d must be refused", gas)
		require.Contains(t, err.Error(), "cannot be negative")
		require.False(t, ok, "gas %d must never report a sufficient fee", gas)
	}

	// And a negative on the other side is refused too.
	ok, err = nodeMin.IsGTE(GasPrice{Gas: -1000, Price: Coin{Denom: "ugnot", Amount: 1}})
	require.Error(t, err)
	require.False(t, ok)
}

func TestGasPriceJSONObjectAtomic(t *testing.T) {
	before := GasPrice{Gas: 1000, Price: Coin{Denom: "ugnot", Amount: 7}}
	for _, tc := range []struct {
		name, input string
		want        GasPrice
		wantErr     bool
	}{
		{"error after gas", `{"gas":1,"price":{"denom":"foo","amount":"bad"}}`, before, true},
		{"error before price", `{"gas":"bad","price":{"denom":"foo","amount":2}}`, before, true},
		{"partial", `{"gas":1}`, GasPrice{Gas: 1, Price: before.Price}, false},
		{"complete", `{"gas":2,"price":{"denom":"ugnot","amount":3}}`, GasPrice{Gas: 2, Price: Coin{Denom: "ugnot", Amount: 3}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := before
			err := json.Unmarshal([]byte(tc.input), &got)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.want, got)
		})
	}
}

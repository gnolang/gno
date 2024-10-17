package std

import (
	"strings"
	"testing"
)

func TestGasPriceGTE(t *testing.T) {
	tests := []struct {
		name        string
		gp          GasPrice
		gpB         GasPrice
		expectPanic bool
		panicMsg    string
		expected    bool // for non-panic cases: whether gp.IsGTE(gpB) should return true or false
	}{
		// Panic cases: Different denominations
		{
			name: "Different Denominations Panic",
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
			expectPanic: true,
			panicMsg:    "gas price denominations should be equal",
		},
		// Panic cases: Zero Gas values
		{
			name: "Zero Gas in gp Panic",
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
			expectPanic: true,
			panicMsg:    "GasPrice.Gas cannot be zero",
		},
		{
			name: "Zero Gas in gpB Panic",
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
			expectPanic: true,
			panicMsg:    "GasPrice.Gas cannot be zero",
		},
		// Valid cases: No panic, just compare gas prices
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
			expectPanic: false,
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
			expectPanic: false,
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
			expectPanic: false,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("Test %s failed: expected no panic, but got panic: %v", tt.name, r)
					} else if tt.expectPanic && r != nil {
						// Check if the panic message contains the expected substring
						panicMsg := r.(string)
						if tt.expectPanic && !strings.Contains(panicMsg, tt.panicMsg) {
							t.Errorf("Test %s failed: expected panic message containing %q, but got %q", tt.name, tt.panicMsg, panicMsg)
						}
					}
				} else if tt.expectPanic {
					t.Errorf("Test %s failed: expected panic, but no panic occurred", tt.name)
				}
			}()

			if !tt.expectPanic {
				got := tt.gp.IsGTE(tt.gpB)
				if got != tt.expected {
					t.Errorf("Test %s failed: expected result %v, got %v", tt.name, tt.expected, got)
				}
			} else {
				// This will panic, but we handle it in the defer/recover above.
				_ = tt.gp.IsGTE(tt.gpB)
			}
		})
	}
}

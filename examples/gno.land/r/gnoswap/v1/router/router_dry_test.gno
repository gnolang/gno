package router

import (
	"testing"

	"gno.land/p/demo/uassert"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestProcessResult(t *testing.T) {
	tests := []struct {
		name            string
		swapType        SwapType
		resultAmountIn  string
		resultAmountOut string
		amountSpecified string
		expected        string
	}{
		{
			name:            "ExactIn - Normal",
			swapType:        ExactIn,
			resultAmountIn:  "100",
			resultAmountOut: "95",
			amountSpecified: "100",
			expected:        "95",
		},
		{
			name:            "ExactIn - Input Mismatch",
			swapType:        ExactIn,
			resultAmountIn:  "99",
			resultAmountOut: "5",
			amountSpecified: "100",
			expected:        "-1",
		},
		{
			name:            "ExactOut - Normal",
			swapType:        ExactOut,
			resultAmountIn:  "105",
			resultAmountOut: "100",
			amountSpecified: "100",
			expected:        "105",
		},
		{
			name:            "ExactOut - Output Mismatch",
			swapType:        ExactOut,
			resultAmountIn:  "105",
			resultAmountOut: "95",
			amountSpecified: "100",
			expected:        "-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultAmountIn, _ := u256.FromDecimal(tt.resultAmountIn)
			resultAmountOut, _ := u256.FromDecimal(tt.resultAmountOut)
			amountSpecified, _ := i256.FromDecimal(tt.amountSpecified)
			amountLimit, _ := i256.FromDecimal("10")
			if tt.swapType == ExactOut {
				amountLimit, _ = i256.FromDecimal("500")
			}

			result := processResult(tt.swapType, resultAmountIn, resultAmountOut, amountSpecified, amountLimit)
			uassert.Equal(t, result, tt.expected)
		})
	}
}

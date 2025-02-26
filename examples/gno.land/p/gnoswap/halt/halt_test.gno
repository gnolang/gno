package halt

import (
	"testing"
)

func TestOpTypeFromString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected OpType
	}{
		{
			name:     "default type: swap",
			input:    rawSwap,
			expected: OpTypeSwap,
		},
		{
			name:     "default type: liquidity",
			input:    rawLiquidity,
			expected: OpTypeLiquidity,
		},
		{
			name:     "default type: withdraw",
			input:    rawWithdraw,
			expected: OpTypeWithdraw,
		},
		{
			name:     "default type: governance",
			input:    rawGovernance,
			expected: OpTypeGovernance,
		},
		{
			name:     "custom type",
			input:    "custom_operation",
			expected: OpType("custom_operation"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := OpTypeFromString(tt.input)
			if err != nil {
				t.Errorf("OpTypeFromString() error = %v", err)
				return
			}
			if result != tt.expected {
				t.Errorf("OpTypeFromString() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

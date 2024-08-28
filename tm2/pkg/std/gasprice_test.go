package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGasPrice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected GasPrice
		hasError bool
	}{
		{
			name:     "Valid case 10foo/1gas",
			input:    "10foo/1gas",
			expected: GasPrice{Gas: 1, Price: Coin{Denom: "foo", Amount: 10}},
			hasError: false,
		},
		{
			name:     "Valid case 100bar/2gas",
			input:    "100bar/2gas",
			expected: GasPrice{Gas: 2, Price: Coin{Denom: "bar", Amount: 100}},
			hasError: false,
		},
		{
			name:     "Valid case 5coin/0gas",
			input:    "5coin/0gas",
			expected: GasPrice{Gas: 0, Price: Coin{Denom: "coin", Amount: 5}},
			hasError: false,
		},
		{
			name:     "Invalid case",
			input:    "invalid",
			expected: GasPrice{},
			hasError: true,
		},
		{
			name:     "Invalid denom case",
			input:    "10foo/1coin",
			expected: GasPrice{},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseGasPrice(tc.input)
			if tc.hasError {
				assert.True(t, err != nil, "expected error but got none")
			} else {
				assert.Nil(t, err, "unexpected error")
				assert.Equal(t, result, tc.expected)
			}
		})
	}
}

func TestParseGasPrices(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected []GasPrice
		hasError bool
	}{
		{
			name:  "Valid case 10foo/1gas and 5bar/2gas",
			input: "10foo/1gas;5bar/2gas",
			expected: []GasPrice{
				{Gas: 1, Price: Coin{Denom: "foo", Amount: 10}},
				{Gas: 2, Price: Coin{Denom: "bar", Amount: 5}},
			},
			hasError: false,
		},
		{
			name:  "Valid case 5coin/0gas",
			input: "5coin/0gas",
			expected: []GasPrice{
				{Gas: 0, Price: Coin{Denom: "coin", Amount: 5}},
			},
			hasError: false,
		},
		{
			name:     "Invalid case",
			input:    "invalid",
			expected: nil,
			hasError: true,
		},
		{
			name:     "Invalid denom case 10foo/1coin and 5bar/2gas",
			input:    "10foo/1coin;5bar/2gas",
			expected: nil,
			hasError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseGasPrices(tc.input)
			if tc.hasError {
				assert.True(t, err != nil, "expected error but got none")
			} else {
				assert.Nil(t, err, "unexpected error")
				assert.Equal(t, result, tc.expected)
			}
		})
	}
}

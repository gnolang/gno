package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGasPrice(t *testing.T) {
	tests := []struct {
		input    string
		expected GasPrice
		hasError bool
	}{
		{
			input:    "10foo/1gas",
			expected: GasPrice{Gas: 1, Price: Coin{Denom: "foo", Amount: 10}},
			hasError: false,
		},
		{
			input:    "100bar/2gas",
			expected: GasPrice{Gas: 2, Price: Coin{Denom: "bar", Amount: 100}},
			hasError: false,
		},
		{
			input:    "5coin/0gas",
			expected: GasPrice{Gas: 0, Price: Coin{Denom: "coin", Amount: 5}},
			hasError: false,
		},
		{
			input:    "invalid",
			expected: GasPrice{},
			hasError: true,
		},
		{
			input:    "10foo/1coin",
			expected: GasPrice{},
			hasError: true,
		},
	}

	for _, test := range tests {
		result, err := ParseGasPrice(test.input)
		if test.hasError {
			assert.True(t, err != nil, "expected error but got none")
		} else {
			assert.Nil(t, err, "unexpected error")
			assert.Equal(t, result, test.expected)
		}
	}
}

func TestParseGasPrices(t *testing.T) {
	tests := []struct {
		input    string
		expected []GasPrice
		hasError bool
	}{
		{
			input: "10foo/1gas;5bar/2gas",
			expected: []GasPrice{
				{Gas: 1, Price: Coin{Denom: "foo", Amount: 10}},
				{Gas: 2, Price: Coin{Denom: "bar", Amount: 5}},
			},
			hasError: false,
		},
		{
			input: "5coin/0gas",
			expected: []GasPrice{
				{Gas: 0, Price: Coin{Denom: "coin", Amount: 5}},
			},
			hasError: false,
		},
		{
			input:    "invalid",
			expected: nil,
			hasError: true,
		},
		{
			input:    "10foo/1coin;5bar/2gas",
			expected: nil,
			hasError: true,
		},
	}

	for _, test := range tests {
		result, err := ParseGasPrices(test.input)
		if test.hasError {
			assert.True(t, err != nil, "expected error but got none")
		} else {
			assert.Nil(t, err, "unexpected error")
			assert.Equal(t, result, test.expected)
		}
	}
}

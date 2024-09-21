package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGasPrice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          string
		expectedResult GasPrice
		expectedError  string
	}{
		{
			name:  "valid gas price",
			input: "1000ugnot/500gas",
			expectedResult: GasPrice{
				Gas: 500,
				Price: Coin{
					Denom:  "ugnot",
					Amount: 1000,
				},
			},
			expectedError: "",
		},
		{
			name:           "invalid gas price format",
			input:          "1000ugnot500gas",
			expectedResult: GasPrice{},
			expectedError:  "invalid gas price: 1000ugnot500gas",
		},
		{
			name:           "invalid price",
			input:          "invalidprice/500gas",
			expectedResult: GasPrice{},
			expectedError:  "invalid coin expression: invalidprice",
		},
		{
			name:           "invalid gas denom",
			input:          "1000ugnot/invalidgas",
			expectedResult: GasPrice{},
			expectedError:  "invalid coin expression: invalidgas",
		},
		{
			name:           "invalid gas denom",
			input:          "1000ugnot/1000ugnot",
			expectedResult: GasPrice{},
			expectedError:  "(invalid gas denom)",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseGasPrice(tc.input)
			assert.Equal(t, tc.expectedResult, result)
			if err != nil {
				assert.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}

func TestParseGasPrices(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          string
		expectedResult []GasPrice
		expectedError  string
	}{
		{
			name:  "Valid case 10foo/1gas and 5bar/2gas",
			input: "10foo/1gas;5bar/2gas",
			expectedResult: []GasPrice{
				{Gas: 1, Price: Coin{Denom: "foo", Amount: 10}},
				{Gas: 2, Price: Coin{Denom: "bar", Amount: 5}},
			},
			expectedError: "",
		},
		{
			name:  "Valid case 5coin/0gas",
			input: "5coin/0gas",
			expectedResult: []GasPrice{
				{Gas: 0, Price: Coin{Denom: "coin", Amount: 5}},
			},
			expectedError: "",
		},
		{
			name:           "Invalid gas price",
			input:          "invalidgas",
			expectedResult: nil,
			expectedError:  "invalid gas price: invalidgas",
		},
		{
			name:           "Invalid denom case 10foo/1coin and 5bar/2gas",
			input:          "10foo/1coin;5bar/2gas",
			expectedResult: nil,
			expectedError:  "(invalid gas denom)",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := ParseGasPrices(tc.input)
			assert.Equal(t, tc.expectedResult, result)
			if err != nil {
				assert.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}

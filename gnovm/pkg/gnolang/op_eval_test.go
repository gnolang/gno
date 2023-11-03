package gnolang_test

import (
	"fmt"
	"math/big"
	"testing"
)

// parseIntegerLiteral function is a Mock-up of the doOpEval() function in `op_eval.go`.
func parseIntegerLiteral(value string) (*big.Int, error) {
	if len(value) > 2 && value[0] == '0' {
		bi := big.NewInt(0)
		var ok bool
		switch value[1] {
		case 'b', 'B':
			_, ok = bi.SetString(value[2:], 2)
		case 'o', 'O':
			_, ok = bi.SetString(value[2:], 8)
		case 'x', 'X':
			_, ok = bi.SetString(value[2:], 16)
		case '0', '1', '2', '3', '4', '5', '6', '7':
			_, ok = bi.SetString(value, 8)
		default:
			ok = false
		}
		if !ok {
			return nil, fmt.Errorf("invalid integer constant: %s", value)
		}
		return bi, nil
	} else {
		bi, ok := new(big.Int).SetString(value, 10)
		if !ok {
			return nil, fmt.Errorf("invalid integer constant: %s", value)
		}
		return bi, nil
	}
}

type testCase struct {
	input    string
	expected string
	hasError bool
}

func TestParseIntegerLiteral(t *testing.T) {
	testCases := []testCase{
		{"012345", "5349", false},
		{"02001", "1025", false},
		{"002001", "1025", false},
		{"0002001", "1025", false},
		{"0o12345", "5349", false},
		{"0x1a2b", "6699", false},
		{"0xbeefcafe", "3203386110", false},
		{"0b1010", "10", false},
		{"12345", "12345", false},
		{"invalid", "", true},
		{"0o12345invalid", "", true},
		{"", "", true},
	}

	for _, tc := range testCases {
		result, err := parseIntegerLiteral(tc.input)
		if tc.hasError {
			if err == nil {
				t.Errorf("Expected error for input %s, but got none", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", tc.input, err)
			} else if result.String() != tc.expected {
				t.Errorf("For input %s, expected %s but got %s", tc.input, tc.expected, result.String())
			}
		}
	}
}

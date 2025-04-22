package evm

import (
	"errors"
	"fmt"
	"testing"

	"gno.land/p/demo/evm/uint256"
)

func TestExecuteEvm(t *testing.T) {
	testCases := []struct {
		name               string
		code               []byte
		expected_success   []byte
		expected_stackdata []*uint256.Uint
		expected_error     error
	}{
		{
			name:               "simple addition",
			code:               []byte{0x60, 1, 0x60, 1, 0x01, 0x00}, // PUSH1 1, PUSH1 1, ADD
			expected_stackdata: []*uint256.Uint{uint256.NewFrom64s(2)},
			expected_error:     nil,
		},
		{
			name:               "error can't add on empty stack",
			code:               []byte{0x01}, // ADD
			expected_stackdata: []*uint256.Uint{},
			expected_error:     errors.New("can't ADD: empty stack"),
		},
		{
			name:               "error push out of bounds",
			code:               []byte{0x60}, // PUSH1
			expected_stackdata: []*uint256.Uint{},
			expected_error:     outOfBounds,
		},
		{
			name:               "simple addition and mul",
			code:               []byte{byte(PUSH1), 10, byte(PUSH1), 11, byte(ADD), byte(PUSH1), 2, byte(MUL), byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_stackdata: []*uint256.Uint{uint256.NewFrom64s(42)},
			expected_error:     nil,
		},
		{
			name:               "exchange 1st and 2nd stack items",
			code:               []byte{byte(PUSH0), byte(PUSH1), 0x42, byte(SWAP1), byte(STOP)}, // PUSH0, PUSH1 0x42, SWAP1
			expected_stackdata: []*uint256.Uint{uint256.NewFrom64s(66), uint256.New()},
			expected_error:     nil,
		},
		{
			name:               "push2",
			code:               []byte{byte(PUSH2), 0x42, 0x42, byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_stackdata: []*uint256.Uint{uint256.NewFrom64s(16962)},
			expected_error:     nil,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			st := newState()
			err := st.executeByteCode(tc.code)
			result := st.getStackData()
			if (err != nil && tc.expected_error != nil && (err.Error() != tc.expected_error.Error())) || (err == nil && tc.expected_error != nil) {
				t.Errorf("I expected this error:\n -> '%s'\nbut I got that one:\n -> '%s'", tc.expected_error, err)
			}
			for i := 0; i < len(result); i++ {
				if !result[i].Equals(tc.expected_stackdata[i]) {
					t.Errorf("I expected this output:\n -> '%s'\nbut I got that one:\n -> '%s'", tc.expected_stackdata[i].HexString(), result[i].HexString())
					return
				}
			}
		})
	}
}

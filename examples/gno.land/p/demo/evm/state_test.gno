package evm

import (
	"errors"
	"fmt"
	"testing"
)

func TestExecuteEvm(t *testing.T) {
	testCases := []struct {
		name             string
		code             []byte
		expected_success []byte
		expected_error   error
	}{
		{
			name:             "simple addition",
			code:             []byte{0x60, 1, 0x60, 1, 0x01, 0x00}, // PUSH1 1, PUSH1 1, ADD
			expected_success: []byte{2},
			expected_error:   nil,
		},
		{
			name:             "error can't add on empty stack",
			code:             []byte{0x01}, // ADD
			expected_success: []byte{},
			expected_error:   errors.New("can't ADD: empty stack"),
		},
		{
			name:             "error push out of bounds",
			code:             []byte{0x60}, // PUSH1
			expected_success: []byte{},
			expected_error:   outOfBounds,
		},
		{
			name:             "simple addition and mul",
			code:             []byte{byte(PUSH1), 10, byte(PUSH1), 11, byte(ADD), byte(PUSH1), 2, byte(MUL), byte(STOP)}, // PUSH1 10, PUSH1 11, ADD, PUSH1 2, MUL
			expected_success: []byte{42},
			expected_error:   nil,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			st := newState()
			err := st.executeByteCode(tc.code)
			result := st.getResult()
			if (err != nil && tc.expected_error != nil && (err.Error() != tc.expected_error.Error())) || (err == nil && tc.expected_error != nil) {
				t.Errorf("I expected this error:\n -> '%s'\nbut I got than one:\n -> '%s'", tc.expected_error, err)
			}
			for i := 0; i < len(result); i++ {
				if result[i] != tc.expected_success[i] {
					t.Errorf("I expected this output:\n -> '%s'\nbut I got than one:\n -> '%s'", tc.expected_success, result)
					return
				}
			}
		})
	}
}

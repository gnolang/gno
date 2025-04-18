package evm

import "testing"

func TestExecuteEvm(t *testing.T) {
	testCases := []struct {
		name             string
		code             []byte
		expected_success []byte
		expected_error   error
	}{
		{
			name:             "error no return",
			code:             []byte{0x60, 0x0a, 0x60, 0x0b, 0x01}, // PUSH1 0x0a, PUSH1 0x0b, ADD
			expected_success: []byte{},
			expected_error:   nil,
		},
		{
			name:             "error can't add on empty stack",
			code:             []byte{0x0a}, // ADD
			expected_success: []byte{},
			expected_error:   nil,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Execute(tc.code)
			if err != tc.expected_error {
				t.Errorf("Expected success output: %x, but got: %x", tc.expected_success, result)
			}
			for i := 0; i < len(result); i++ {
				if result[i] != tc.expected_success[i] {
					t.Errorf("Expected success output: %x, but got: %x", tc.expected_success, result)
					return
				}
			}
		})
	}
}

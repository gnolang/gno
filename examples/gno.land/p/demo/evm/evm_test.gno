package evm

import "testing"

func TestExecuteEvm(t *testing.T) {
	testCases := []struct {
		name     string
		code     []byte
		expected []byte
	}{
		{
			name:     "return 0x0a + 0x0b",
			code:     []byte{0x60, 0x0a, 0x60, 0x0b, 0x01}, // PUSH1 0x0a, PUSH1 0x0b, ADD
			expected: []byte{0x0a + 0x0b},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Execute(tc.code)
			for i := 0; i < len(result); i++ {
				if result[i] != tc.expected[i] {
					t.Errorf("Expected output: %x, but got: %x", tc.expected, result)
					return
				}
			}
		})
	}
}

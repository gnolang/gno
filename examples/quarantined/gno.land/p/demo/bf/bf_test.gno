package bf

import "testing"

func TestExecuteBrainfuck(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "hello",
			code:     "++++++++++[>+++++++>++++++++++>+++>+<<<<-]>++.>+.+++++++..+++.>++.<<+++++++++++++++.>.+++.------.--------.",
			expected: "Hello World",
		},
		{
			name:     "increment",
			code:     "+++++ +++++ [ > +++++ ++ < - ] > +++++ .",
			expected: "K",
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Execute(tc.code)
			if result != tc.expected {
				t.Errorf("Expected output: %s, but got: %s", tc.expected, result)
			}
		})
	}
}

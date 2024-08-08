package users

import (
	"testing"
)

func TestAddressOrNameValidation(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "Valid address",
			address:  "g18syxa0vh0vmne90mwhtynjet0zgeqf6prh3ryy",
			expected: true,
		},
		{
			name:     "Invalid address",
			address:  "0x123456789012345678901234567890123456789",
			expected: false,
		},
		{
			name:     "Valid name",
			address:  "@test",
			expected: true,
		},
		{
			name:     "Invalid name",
			address:  "test@",
			expected: false,
		},
		{
			name:     "Empty",
			address:  "",
			expected: false,
		},
		{
			name:     "Only @",
			address:  "@",
			expected: false,
		},
		{
			name:     "One letter",
			address:  "@a",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AddressOrName(test.address).IsValid()
			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

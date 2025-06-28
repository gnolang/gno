package gnolang

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func TestDerivePkgCryptoAddr(t *testing.T) {
	validAddr := "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	cryptoAddr, err := crypto.AddressFromBech32(validAddr)
	if err != nil {
		t.Fatalf("failed to parse bech32 address: %v", err)
	}
	tests := []struct {
		name     string
		pkgPath  string
		expected crypto.Address
	}{
		{
			name:     "new ephemeral run path",
			pkgPath:  "gno.land/e/" + validAddr + "/run",
			expected: cryptoAddr,
		},
		{
			name:     "old run path",
			pkgPath:  "gno.land/r/" + validAddr + "/run",
			expected: crypto.AddressFromPreimage([]byte("pkgPath:gno.land/r/" + validAddr + "/run")),
		},
		{
			name:     "regular realm path with address as namespace",
			pkgPath:  "gno.land/r/" + validAddr + "/test",
			expected: crypto.AddressFromPreimage([]byte("pkgPath:gno.land/r/" + validAddr + "/test")),
		},
		{
			name:     "regular realm path with username as namespace",
			pkgPath:  "gno.land/r/foobar/test",
			expected: crypto.AddressFromPreimage([]byte("pkgPath:gno.land/r/foobar/test")),
		},
		{
			name:     "ephemeral path",
			pkgPath:  "gno.land/e/" + validAddr + "/test",
			expected: crypto.AddressFromPreimage([]byte("pkgPath:gno.land/e/" + validAddr + "/test")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DerivePkgCryptoAddr(tt.pkgPath)
			if result != tt.expected {
				t.Errorf("DerivePkgCryptoAddr(%q) = %v, want %v", tt.pkgPath, result, tt.expected)
			}
		})
	}
}

func TestDerivePkgBech32Addr(t *testing.T) {
	validAddr := "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	tests := []struct {
		name     string
		pkgPath  string
		expected crypto.Bech32Address
	}{
		{
			name:     "new ephemeral run path",
			pkgPath:  "gno.land/e/" + validAddr + "/run",
			expected: crypto.Bech32Address(validAddr),
		},
		{
			name:     "old run path",
			pkgPath:  "gno.land/r/" + validAddr + "/run",
			expected: crypto.AddressFromPreimage([]byte("pkgPath:gno.land/r/" + validAddr + "/run")).Bech32(),
		},
		{
			name:     "regular realm path",
			pkgPath:  "gno.land/r/" + validAddr + "/test",
			expected: crypto.AddressFromPreimage([]byte("pkgPath:gno.land/r/" + validAddr + "/test")).Bech32(),
		},
		{
			name:     "ephemeral path",
			pkgPath:  "gno.land/e/" + validAddr + "/test",
			expected: crypto.AddressFromPreimage([]byte("pkgPath:gno.land/e/" + validAddr + "/test")).Bech32(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DerivePkgBech32Addr(tt.pkgPath)
			if result != tt.expected {
				t.Errorf("DerivePkgBech32Addr(%q) = %v, want %v", tt.pkgPath, result, tt.expected)
			}
		})
	}
}

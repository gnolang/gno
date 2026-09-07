package gnolang

import (
	"testing"

	"github.com/stretchr/testify/require"

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

func TestObjectIDDerivePath(t *testing.T) {
	t.Parallel()

	var (
		pkgID = PkgIDFromPkgPath("gno.land/r/demo/objectid")
		other = PkgIDFromPkgPath("gno.land/r/demo/objectid_other")
	)

	tests := []struct {
		name string
		oid  ObjectID
		want string
	}{
		{
			name: "zero id has nothing to derive from",
			oid:  ObjectID{},
			want: "",
		},
		{
			// PkgID is stamped at allocation, NewTime only at finalization:
			// an object that was never persisted has no address.
			name: "allocated but unfinalized",
			oid:  ObjectID{PkgID: pkgID},
			want: "",
		},
		{
			name: "finalized",
			oid:  ObjectID{PkgID: pkgID, NewTime: 7},
			want: DeriveObjectIDCryptoAddr(ObjectID{PkgID: pkgID, NewTime: 7}).String(),
		},
		{
			name: "another tick of the same realm",
			oid:  ObjectID{PkgID: pkgID, NewTime: 8},
			want: DeriveObjectIDCryptoAddr(ObjectID{PkgID: pkgID, NewTime: 8}).String(),
		},
		{
			name: "same tick of another realm",
			oid:  ObjectID{PkgID: other, NewTime: 7},
			want: DeriveObjectIDCryptoAddr(ObjectID{PkgID: other, NewTime: 7}).String(),
		},
	}

	derived := make(map[string]string, len(tests))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.oid.DerivePath()
			require.Equal(t, tt.want, got)
			// Deriving twice must not move.
			require.Equal(t, got, tt.oid.DerivePath())
			if got == "" {
				return
			}
			// Both halves take part, so no two ids share an address, and none
			// collides with the address its realm derives from its pkgpath.
			require.NotContains(t, derived, got, "address collision with %q", derived[got])
			derived[got] = tt.name
			require.NotEqual(t, DerivePkgBech32Addr("gno.land/r/demo/objectid").String(), got)
		})
	}
}

func TestDeriveObjectIDCryptoAddrRejectsIncompleteIDs(t *testing.T) {
	t.Parallel()

	pkgID := PkgIDFromPkgPath("gno.land/r/demo/objectid")

	tests := []struct {
		name string
		oid  ObjectID
	}{
		{name: "zero", oid: ObjectID{}},
		{name: "no pkgID", oid: ObjectID{NewTime: 7}},
		{name: "no newTime", oid: ObjectID{PkgID: pkgID}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Panics(t, func() { DeriveObjectIDCryptoAddr(tt.oid) })
		})
	}
}

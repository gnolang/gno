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

func TestObjectIDDeriveAddress(t *testing.T) {
	t.Parallel()

	// PkgID is the first input to the address. Pin it as a literal so a change
	// to PkgIDFromPkgPath, which would move every object address in the realm,
	// is reported here rather than passing silently.
	const (
		pkgIDStr   = "RID0096D59EE52BF51629778FC5525C67BE07EC5133"
		otherIDStr = "RID009F60F0E677E06D78B4ECFDE9638F1A1000DD3F"
	)
	pkgID := PkgIDFromPkgPath("gno.land/r/demo/objectid")
	other := PkgIDFromPkgPath("gno.land/r/demo/objectid_other")
	require.Equal(t, pkgIDStr, pkgID.String())
	require.Equal(t, otherIDStr, other.String())

	tests := []struct {
		name string
		oid  ObjectID
		// preimage, when set, is spelled out independently of the code under
		// test so a changed prefix or separator reports as a layout change
		// rather than a hash miss.
		preimage string
		want     string
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
			name:     "finalized",
			oid:      ObjectID{PkgID: pkgID, NewTime: 7},
			preimage: "objectid:" + pkgIDStr + ":7",
			want:     "g10qdrafyefex7t0nn9ru2sxge00hlm58p0mpvqd",
		},
		{
			name:     "another tick of the same realm",
			oid:      ObjectID{PkgID: pkgID, NewTime: 8},
			preimage: "objectid:" + pkgIDStr + ":8",
			want:     "g19qtnngqw7pjvfz9lpc4s464ptz8qtpnqr6wpkg",
		},
		{
			name:     "same tick of another realm",
			oid:      ObjectID{PkgID: other, NewTime: 7},
			preimage: "objectid:" + otherIDStr + ":7",
			want:     "g1vfgnmjj62j2ge6c5en9w7kmat56nysd6nst27t",
		},
	}

	// Both halves take part, so no two ids share an address, and none collides
	// with the address its realm derives from its pkgpath. Checked here rather
	// than in the subtests, which run in parallel and share nothing.
	derived := make(map[string]string, len(tests))
	for _, tt := range tests {
		got := tt.oid.DeriveAddress()
		if got == "" {
			continue
		}
		require.NotContains(t, derived, got, "address collision with %q", derived[got])
		require.NotEqual(t, DerivePkgBech32Addr("gno.land/r/demo/objectid").String(), got)
		derived[got] = tt.name
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.oid.DeriveAddress()
			require.Equal(t, tt.want, got)
			// Deriving twice must not move.
			require.Equal(t, got, tt.oid.DeriveAddress())

			if tt.preimage != "" {
				require.Equal(t, tt.want, crypto.AddressFromPreimage([]byte(tt.preimage)).String())
			}
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
			t.Parallel()

			require.Panics(t, func() { DeriveObjectIDCryptoAddr(tt.oid) })
		})
	}
}

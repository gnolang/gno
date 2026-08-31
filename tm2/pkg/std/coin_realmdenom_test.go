package std

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// ParseRealmDenom decides whether a stored realm-shaped denom is one a realm
// could have issued. Its only caller is BalanceKeysInvariant, which reports the
// ones that are not as genesis-authored or corrupt -- so every branch here is a
// sentence that invariant may have to print, and none of them were exercised.
func TestParseRealmDenomBranches(t *testing.T) {
	t.Parallel()

	longPath := "gno.land/r/" + strings.Repeat("a", pkgPathLimit)
	longBase := strings.Repeat("a", maxBaseDenomLength+1)

	cases := []struct {
		name, denom string
		wantErr     string
		pkgPath     string
		base        string
	}{
		{"plain denom is not realm-qualified", "ugnot", "not realm-qualified", "", ""},
		{"no colon, so no base name", "/gno.land/r/demo/foo", "has no base name", "", ""},
		{"package path over the limit", "/" + longPath + ":gold", "over the", "", ""},
		{"empty base name", "/gno.land/r/demo/foo:", "empty base name", "", ""},
		{"base name over the limit", "/gno.land/r/demo/foo:" + longBase, "over the", "", ""},
		{"base name starting with a digit", "/gno.land/r/demo/foo:1gold", "must start with a-z", "", ""},
		{"base name starting with uppercase", "/gno.land/r/demo/foo:Gold", "must start with a-z", "", ""},
		{"base name with an inner uppercase", "/gno.land/r/demo/foo:goLd", "[a-z][a-z0-9]*", "", ""},
		{"base name with a separator inside", "/gno.land/r/demo/foo:go-ld", "[a-z][a-z0-9]*", "", ""},

		{"ordinary realm denom", "/gno.land/r/demo/foo:gold", "", "gno.land/r/demo/foo", "gold"},
		{"digits after the first letter", "/gno.land/r/demo/foo:g01d", "", "gno.land/r/demo/foo", "g01d"},
		{
			// Only the first colon separates, so one in the base name is still
			// a base-name character and is refused as such.
			"second colon lands in the base name",
			"/gno.land/r/demo/foo:go:ld", "[a-z][a-z0-9]*", "", "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pkgPath, base, err := ParseRealmDenom(tt.denom)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr,
					"the message must name the real problem")
				require.Empty(t, pkgPath)
				require.Empty(t, base)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.pkgPath, pkgPath)
			require.Equal(t, tt.base, base)
		})
	}
}

// The banker refuses a base name under three characters; this does not, and
// that gap is deliberate -- the minimum is issuance ergonomics and says nothing
// about whether stored state is well formed. Pinned so that closing the gap has
// to be a decision rather than a tidy-up: realm denoms already in test fixtures
// do not satisfy the minimum, and enforcing it here would reclassify them as
// corrupt to BalanceKeysInvariant.
func TestParseRealmDenomTakesAShortBaseTheBankerWouldRefuse(t *testing.T) {
	t.Parallel()

	for _, base := range []string{"a", "go"} {
		pkgPath, got, err := ParseRealmDenom("/gno.land/r/demo/foo:" + base)
		require.NoError(t, err, "a %d-character base must still parse", len(base))
		require.Equal(t, "gno.land/r/demo/foo", pkgPath)
		require.Equal(t, base, got)
	}

	// The maximum is mirrored, because that one does bound stored state.
	atLimit := strings.Repeat("a", maxBaseDenomLength)
	_, got, err := ParseRealmDenom("/gno.land/r/demo/foo:" + atLimit)
	require.NoError(t, err, "a base name exactly at the limit must parse")
	require.Equal(t, atLimit, got)
}

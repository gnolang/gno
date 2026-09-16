package gnoland

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tmver "github.com/gnolang/gno/tm2/pkg/version"
)

func TestParseReleaseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		// The current shape.
		{"semver", "v1.2.0", "v1.2.0", true},
		{"semver patch", "v1.2.3", "v1.2.3", true},
		{"semver zero", "v0.0.0", "v0.0.0", true},
		{"semver large", "v10.20.30", "v10.20.30", true},
		{"semver prerelease", "v1.3.0-rc.1", "v1.3.0-rc.1", true},
		{"semver build metadata ignored", "v1.2.0+deadbeef", "v1.2.0", true},
		{"semver prerelease and build", "v1.3.0-rc.1+deadbeef", "v1.3.0-rc.1", true},
		// No int conversion happens, so a component too large for int64 is a
		// version like any other rather than a parse failure.
		{"component beyond int64", "v99999999999999999999.0.0", "v99999999999999999999.0.0", true},

		// Betanet's frozen shape. Both tags that ever used it, exactly as they
		// are compiled into the binaries that ran that chain. They widen to the
		// v-line so that one ordering covers both.
		{"legacy betanet .0", "chain/gnoland1.0", "v1.0.0", true},
		{"legacy betanet .1", "chain/gnoland1.1", "v1.1.0", true},
		{"legacy hypothetical", "chain/gnoland2.3", "v2.3.0", true},

		// Everything a node might actually be running that must NOT parse.
		{"plain go build", "develop", "", false},
		{"off-tag make build", "master.3335+bc43a5fb7", "", false},
		{"unnumbered chain tag", "chain/mainnet", "", false},
		{"unnumbered chain tag pearl", "chain/pearl", "", false},
		{"chain tag stripped by CI", "mainnet", "", false},
		{"empty", "", "", false},

		// Malformed input that must not be coerced into a valid-looking version.
		{"no v prefix", "1.2.0", "", false},
		{"two components", "v1.2", "", false},
		{"one component", "v1", "", false},
		{"four components", "v1.2.3.4", "", false},
		{"non numeric", "v1.x.0", "", false},
		{"empty component", "v1..0", "", false},
		{"signed component", "v1.+2.0", "", false},
		{"negative component", "v1.-2.0", "", false},
		{"leading zero", "v1.02.0", "", false},
		{"leading zero major", "v01.2.0", "", false},
		// A dangling separator would otherwise yield an empty pre-release,
		// which is the sentinel for "this is the final release".
		{"dangling prerelease separator", "v1.2.3-", "", false},
		{"dangling prerelease with build", "v1.2.3-+meta", "", false},
		{"legacy without minor", "chain/gnoland1", "", false},
		{"legacy non numeric", "chain/gnolandX.Y", "", false},
		// The legacy shape gets the same component rules as the v shape. A
		// negative major would otherwise parse and produce a floor that every
		// parseable binary clears — a gate that silently does nothing.
		{"legacy negative major", "chain/gnoland-1.1", "", false},
		{"legacy negative minor", "chain/gnoland1.-1", "", false},
		{"legacy signed major", "chain/gnoland+1.0", "", false},
		{"legacy leading zero", "chain/gnoland01.1", "", false},
		{"legacy three components", "chain/gnoland1.1.1", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseReleaseVersion(tt.in)
			require.Equal(t, tt.ok, ok, "parse ok for %q", tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMeetsMinVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		binary  string
		minimum string
		want    bool
	}{
		// No floor set.
		{"empty minimum accepts anything", "develop", "", true},

		// The regression this shape exists for: before v-tags parsed, a
		// minVersion of "chain/mainnet" degraded to byte equality and refused
		// the very binary the upgrade was cut for, so the chain could not
		// restart. Both directions are now ordered.
		{"newer minor meets floor", "v1.3.0", "v1.2.0", true},
		{"newer patch meets floor", "v1.2.1", "v1.2.0", true},
		{"newer major meets floor", "v2.0.0", "v1.9.9", true},
		{"exact meets floor", "v1.2.0", "v1.2.0", true},
		{"older minor refused", "v1.1.0", "v1.2.0", false},
		{"older patch refused", "v1.2.0", "v1.2.1", false},
		{"older major refused", "v1.9.9", "v2.0.0", false},

		// Pre-releases rank below the release they lead to, so an rc build
		// cannot be mistaken for the release in a halt gate.
		{"prerelease does not meet its release", "v1.3.0-rc.1", "v1.3.0", false},
		{"release meets its prerelease floor", "v1.3.0", "v1.3.0-rc.1", true},
		{"later rc meets earlier rc", "v1.3.0-rc.2", "v1.3.0-rc.1", true},
		{"earlier rc refused", "v1.3.0-rc.1", "v1.3.0-rc.2", false},
		// Numeric pre-release identifiers order numerically, not lexically.
		// Under string comparison rc.2 would clear an rc.10 floor, which is an
		// older binary passing a gate set against a newer one.
		{"rc.10 meets rc.9", "v1.3.0-rc.10", "v1.3.0-rc.9", true},
		{"rc.10 meets rc.2", "v1.3.0-rc.10", "v1.3.0-rc.2", true},
		{"rc.2 refused against rc.10", "v1.3.0-rc.2", "v1.3.0-rc.10", false},
		{"build metadata does not affect ordering", "v1.2.0+abc", "v1.2.0+def", true},

		// Betanet's line and the v-line are one ordering: the v-line continues
		// from where chain/gnoland1.1 left off.
		{"legacy same version", "chain/gnoland1.0", "chain/gnoland1.0", true},
		{"legacy minor ordering", "chain/gnoland1.1", "chain/gnoland1.0", true},
		{"legacy major ordering", "chain/gnoland2.0", "chain/gnoland1.0", true},
		{"legacy older refused", "chain/gnoland1.0", "chain/gnoland1.1", false},
		{"legacy older major refused", "chain/gnoland1.0", "chain/gnoland2.0", false},
		{"legacy binary against empty floor", "chain/gnoland1.0", "", true},
		{"v-line continues past betanet", "v1.2.0", "chain/gnoland1.1", true},
		{"betanet does not satisfy a v-line floor", "chain/gnoland1.1", "v1.2.0", false},

		// An unparseable binary version satisfies no floor. This is what keeps
		// an ad-hoc build off a chain that has gated its restart.
		{"develop refused", "develop", "v1.2.0", false},
		{"off-tag build refused", "master.3335+bc43a5fb7", "v1.2.0", false},
		{"CI-stripped tag refused", "mainnet", "v1.2.0", false},
		{"unnumbered chain tag refused", "chain/mainnet", "v1.2.0", false},

		// An unparseable minVersion still falls back to byte equality, so the
		// pre-existing behaviour is preserved for anything not covered above.
		{"unparseable minimum matches itself", "chain/mainnet", "chain/mainnet", true},
		{"unparseable minimum refuses anything else", "v1.3.0", "chain/mainnet", false},

		// A malformed legacy floor must not parse: if it did, it would order
		// below every real version and admit a stale binary after a halt.
		{"negative legacy floor refuses a stale binary", "v1.2.0", "chain/gnoland-1.3", false},
		{"negative legacy floor refuses a current binary", "v1.3.0", "chain/gnoland-1.3", false},
		// A dangling separator must not compare equal to the release itself.
		{"dangling separator is not the release", "v1.2.3", "v1.2.3-", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, meetsMinVersion(tt.binary, tt.minimum),
				"meetsMinVersion(%q, %q)", tt.binary, tt.minimum)
		})
	}
}

// The two outcomes the release flow exists for, driven through the function a
// starting node actually calls. Everything above tests the parser directly;
// this pins that a real release tag reaches it, because tm2/pkg/version.Version
// is "develop" under `go test` and every other startup test therefore takes the
// byte-equality fallback without exercising the ordering at all.
func TestStartupGateWithAReleaseTag(t *testing.T) {
	orig := tmver.Version
	tmver.Version = "v1.2.0"
	t.Cleanup(func() { tmver.Version = orig })

	t.Run("upgraded binary readmitted after the halt", func(t *testing.T) {
		prmk, ms := newTestParamsKeeper(t, 100, "v1.1.0")
		require.NoError(t, checkNodeStartupParams(prmk, ms, 100, 0))
	})

	t.Run("upgraded binary refused before the halt", func(t *testing.T) {
		prmk, ms := newTestParamsKeeper(t, 100, "v1.1.0")
		require.Error(t, checkNodeStartupParams(prmk, ms, 50, 0))
	})

	t.Run("stale binary refused after the halt", func(t *testing.T) {
		prmk, ms := newTestParamsKeeper(t, 100, "v1.3.0")
		require.Error(t, checkNodeStartupParams(prmk, ms, 100, 0))
	})

	// The regression the v shape was added for: a floor naming the chain tag
	// rather than a release tag falls back to byte equality and refuses the
	// upgraded binary too, so the chain cannot restart at all.
	t.Run("an unparseable floor refuses the upgraded binary", func(t *testing.T) {
		prmk, ms := newTestParamsKeeper(t, 100, "chain/mainnet")
		require.Error(t, checkNodeStartupParams(prmk, ms, 100, 0))
	})
}

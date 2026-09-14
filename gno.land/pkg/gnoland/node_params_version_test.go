package gnoland

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReleaseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want releaseVersion
		ok   bool
	}{
		// The current shape.
		{"semver", "v1.2.0", releaseVersion{major: 1, minor: 2, patch: 0}, true},
		{"semver patch", "v1.2.3", releaseVersion{major: 1, minor: 2, patch: 3}, true},
		{"semver zero", "v0.0.0", releaseVersion{}, true},
		{"semver large", "v10.20.30", releaseVersion{major: 10, minor: 20, patch: 30}, true},
		{"semver prerelease", "v1.3.0-rc.1", releaseVersion{major: 1, minor: 3, pre: "rc.1"}, true},
		{"semver build metadata ignored", "v1.2.0+deadbeef", releaseVersion{major: 1, minor: 2}, true},
		{"semver prerelease and build", "v1.3.0-rc.1+deadbeef", releaseVersion{major: 1, minor: 3, pre: "rc.1"}, true},

		// Betanet's frozen shape. Both tags that ever used it, exactly as they
		// are compiled into the binaries that ran that chain.
		{"legacy betanet .0", "chain/gnoland1.0", releaseVersion{major: 1}, true},
		{"legacy betanet .1", "chain/gnoland1.1", releaseVersion{major: 1, minor: 1}, true},
		{"legacy hypothetical", "chain/gnoland2.3", releaseVersion{major: 2, minor: 3}, true},

		// Everything a node might actually be running that must NOT parse.
		{"plain go build", "develop", releaseVersion{}, false},
		{"off-tag make build", "master.3335+bc43a5fb7", releaseVersion{}, false},
		{"unnumbered chain tag", "chain/mainnet", releaseVersion{}, false},
		{"unnumbered chain tag pearl", "chain/pearl", releaseVersion{}, false},
		{"chain tag stripped by CI", "mainnet", releaseVersion{}, false},
		{"empty", "", releaseVersion{}, false},

		// Malformed input that must not be coerced into a valid-looking version.
		{"no v prefix", "1.2.0", releaseVersion{}, false},
		{"two components", "v1.2", releaseVersion{}, false},
		{"four components", "v1.2.3.4", releaseVersion{}, false},
		{"non numeric", "v1.x.0", releaseVersion{}, false},
		{"empty component", "v1..0", releaseVersion{}, false},
		{"signed component", "v1.+2.0", releaseVersion{}, false},
		{"negative component", "v1.-2.0", releaseVersion{}, false},
		{"leading zero", "v1.02.0", releaseVersion{}, false},
		{"legacy without minor", "chain/gnoland1", releaseVersion{}, false},
		{"legacy non numeric", "chain/gnolandX.Y", releaseVersion{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseReleaseVersion(tt.in)
			require.Equal(t, tt.ok, ok, "parse ok for %q", tt.in)
			if tt.ok {
				assert.Equal(t, tt.want, got)
			}
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, meetsMinVersion(tt.binary, tt.minimum),
				"meetsMinVersion(%q, %q)", tt.binary, tt.minimum)
		})
	}
}

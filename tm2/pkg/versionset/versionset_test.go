package versionset

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CompatibleWith decides whether two nodes may peer at all — p2p/transport.go
// refuses the connection on its error — and RELEASING.md rests a "a MAJOR bump
// partitions the network" claim on it. It carried no tests.
func TestVersionSetCompatibleWith(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ours VersionSet
		them VersionSet
		want VersionSet // nil when an error is expected
	}{
		{
			name: "identical sets negotiate their own major.minor",
			ours: VersionSet{{Name: "bft", Version: "v1.0.0-rc.0"}},
			them: VersionSet{{Name: "bft", Version: "v1.0.0-rc.0"}},
			want: VersionSet{{Name: "bft", Version: "v1.0"}},
		},
		{
			// Patch, pre-release and build take no part: only major.minor is
			// negotiated, so these are the same version to a peer.
			name: "patch and prerelease are discarded",
			ours: VersionSet{{Name: "bft", Version: "v1.2.9-rc.4+meta"}},
			them: VersionSet{{Name: "bft", Version: "v1.2.0"}},
			want: VersionSet{{Name: "bft", Version: "v1.2"}},
		},
		{
			name: "a differing major refuses the peer",
			ours: VersionSet{{Name: "bft", Version: "v1.0.0"}},
			them: VersionSet{{Name: "bft", Version: "v2.0.0"}},
		},
		{
			// The negotiated minor is the lower of the two, in both directions.
			// Before the comparison was fixed it was whichever side was asking.
			name: "a higher minor negotiates down to theirs",
			ours: VersionSet{{Name: "bft", Version: "v1.3.0"}},
			them: VersionSet{{Name: "bft", Version: "v1.1.0"}},
			want: VersionSet{{Name: "bft", Version: "v1.1"}},
		},
		{
			name: "a lower minor negotiates to ours",
			ours: VersionSet{{Name: "bft", Version: "v1.1.0"}},
			them: VersionSet{{Name: "bft", Version: "v1.3.0"}},
			want: VersionSet{{Name: "bft", Version: "v1.1"}},
		},
		{
			name: "an entry only we require refuses the peer",
			ours: VersionSet{{Name: "bft", Version: "v1.0.0"}, {Name: "extra", Version: "v1.0.0"}},
			them: VersionSet{{Name: "bft", Version: "v1.0.0"}},
		},
		{
			name: "an entry only they require refuses us",
			ours: VersionSet{{Name: "bft", Version: "v1.0.0"}},
			them: VersionSet{{Name: "bft", Version: "v1.0.0"}, {Name: "extra", Version: "v1.0.0"}},
		},
		{
			// Accepted, but not negotiated: an entry only one side carries is
			// left out of the result even when it is optional.
			name: "an optional entry missing on one side is fine",
			ours: VersionSet{{Name: "bft", Version: "v1.0.0"}, {Name: "extra", Version: "v1.0.0", Optional: true}},
			them: VersionSet{{Name: "bft", Version: "v1.0.0"}},
			want: VersionSet{{Name: "bft", Version: "v1.0"}},
		},
		{
			name: "an optional entry missing on their side is fine",
			ours: VersionSet{{Name: "bft", Version: "v1.0.0"}},
			them: VersionSet{{Name: "bft", Version: "v1.0.0"}, {Name: "extra", Version: "v1.0.0", Optional: true}},
			want: VersionSet{{Name: "bft", Version: "v1.0"}},
		},
		{
			// semver.Major("") is "", which equals no real major, so a peer
			// advertising nothing is refused rather than waved through.
			name: "an empty version on one side refuses the peer",
			ours: VersionSet{{Name: "bft", Version: "v1.0.0"}},
			them: VersionSet{{Name: "bft", Version: ""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.ours.CompatibleWith(tt.them)
			if tt.want == nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

// The failure mode TestVersionSetIsComplete in tm2/pkg/bft/version guards
// against: if the entry were dropped from *our* set too, semver.Major returns
// "" on both sides, the majors compare equal, and the component is negotiated
// at the empty version — the check passes while checking nothing. One side
// empty is refused (see the table above); it takes both.
func TestVersionSetEmptyOnBothSidesNegotiatesNothing(t *testing.T) {
	t.Parallel()

	ours := VersionSet{{Name: "bft", Version: ""}}
	them := VersionSet{{Name: "bft", Version: ""}}

	res, err := ours.CompatibleWith(them)
	require.NoError(t, err)
	assert.Equal(t, VersionSet{{Name: "bft", Version: ""}}, res)
}

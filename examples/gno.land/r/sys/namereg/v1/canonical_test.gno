package namereg

import (
	"testing"

	"gno.land/p/nt/uassert/v0"

	susers "gno.land/r/sys/users"
)

// Canonicalize is now a delegating shim to susers.Canonicalize; verify
// the local export and the upstream produce the same output for the
// rules namereg/v1 cares about.
func TestCanonicalize_DelegatesToSusers(t *testing.T) {
	cases := []string{
		"",
		"a",
		"alice",
		"vitalik",
		"vital1k",
		"balloon",
		"already-canonical",
		"xyz123",
		"nym-foolbar000",
		"nym-foolbar001",
	}
	for _, in := range cases {
		uassert.Equal(t, susers.Canonicalize(in), Canonicalize(in),
			"namereg.Canonicalize must match susers.Canonicalize")
	}
}

func TestReservedSet_BuiltFromBlacklist(t *testing.T) {
	// Sanity: every entry from reservedNames AND its `+s` form must be
	// in the canonicalized lookup.
	for _, n := range reservedNames {
		_, gotBare := reservedSet[Canonicalize(n)]
		uassert.True(t, gotBare,
			"reservedSet missing canonical(%q)", n)

		_, gotPlural := reservedSet[Canonicalize(n+"s")]
		uassert.True(t, gotPlural,
			"reservedSet missing canonical(%q+s)", n)
	}
}

func TestIsReserved_LiteralAndCanonical(t *testing.T) {
	// Literal entry must be reserved.
	uassert.True(t, IsReserved("admin"), "admin should be reserved")

	// `blogs` covered via `blog`+s rule (blog is in reservedNames).
	uassert.True(t, IsReserved("blogs"), "blogs covered by blog+s")

	// l-substituted: `blog` canonicalizes to `biog`. Both forms blocked
	// because we canonicalize the candidate before lookup.
	uassert.True(t, IsReserved("biog"), "canonical(blog)=biog also blocked")

	// 1-substituted under the new susers.Canonicalize rules: `1` → `i`.
	// `b1og` canonicalizes to `biog`, same as `blog`. Confirms the shim
	// picks up the broader rule set.
	uassert.True(t, IsReserved("b1og"), "1→i variant also blocked")

	// Non-reserved stem should not be flagged.
	uassert.False(t, IsReserved("zulufoxtrot"), "zulufoxtrot is not reserved")

	// Identity passthrough.
	uassert.True(t, IsReserved("tendermint"))
}

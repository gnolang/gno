package namereg

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestCanonicalize(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// Identity: no `l` in input.
		{"", ""},
		{"a", "a"},
		{"alice", "aiice"},          // single l
		{"vitalik", "vitaiik"},      // single l mid
		{"balloon", "baiioon"},      // double l
		{"all", "aii"},              // multiple l in a row
		{"hello", "heiio"},          // l followed by l
		{"already-canonical", "aiready-canonicai"},
		{"abcdef", "abcdef"},        // no l, identity
		{"xyz123", "xyz123"},        // digits and other letters preserved
		{"l", "i"},                  // single-char l
		{"ll", "ii"},                // two l
	}

	for _, c := range cases {
		got := Canonicalize(c.in)
		uassert.Equal(t, c.want, got,
			"Canonicalize(%q) = %q, want %q", c.in, got, c.want)
	}
}

func TestCanonicalize_StaysInLowercaseAlphabet(t *testing.T) {
	// Contract: input is [a-z], output is [a-z]. Verify that Canonicalize
	// never emits anything outside that range for any single-letter input.
	for ch := byte('a'); ch <= byte('z'); ch++ {
		got := Canonicalize(string(ch))
		uassert.Equal(t, 1, len(got), "single char in must give single char out")
		out := got[0]
		if out < 'a' || out > 'z' {
			t.Errorf("Canonicalize(%q) = %q, output not in [a-z]", string(ch), got)
		}
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

	// l-substituted variant (admin has no l, identity) is reserved too.
	uassert.True(t, IsReserved("admin"), "canonical of admin")

	// `blogs` covered via `blog`+s rule (blog is in reservedNames).
	uassert.True(t, IsReserved("blogs"), "blogs covered by blog+s")

	// l-substituted: `blog` canonicalizes to `biog`. Both forms blocked
	// because we canonicalize the candidate before lookup.
	uassert.True(t, IsReserved("biog"), "canonical(blog)=biog also blocked")

	// Non-reserved stem should not be flagged.
	uassert.False(t, IsReserved("zulufoxtrot"), "zulufoxtrot is not reserved")

	// Canonical-form interaction: if `tendermint` is reserved (no l),
	// then a hypothetical l-bearing variant should also be caught via
	// canonicalization. tendermint has no l, so this is identity.
	uassert.True(t, IsReserved("tendermint"))
}

func TestIsStemTaken_DirectStoreOps(t *testing.T) {
	// Exercise the store API in isolation from the Register() chain.
	// Direct write to the store, then verify the lookup sees it.
	canonicalStore.Set("zzdirect", "zzdirect-original")
	defer canonicalStore.Remove("zzdirect")

	// IsStemTaken canonicalizes internally — passing the raw stem
	// (no `l` here, so canonical == raw) finds it.
	existing, taken := IsStemTaken("zzdirect")
	uassert.True(t, taken, "should report taken after direct Set")
	uassert.Equal(t, "zzdirect-original", existing,
		"existing should be the original stem stored as the value")

	_, taken2 := IsStemTaken("zzunknown")
	uassert.False(t, taken2, "unknown stem should report not taken")
}

func TestIsStemTaken_CanonicalizesInternally(t *testing.T) {
	// Footgun guard: the function must canonicalize the input itself.
	// Store is keyed by canonical form; raw and l-substituted variants
	// must both resolve to the same hit.
	canonical := Canonicalize("balloon") // "baiioon"
	canonicalStore.Set(canonical, "balloon")
	defer canonicalStore.Remove(canonical)

	// Raw `balloon` finds itself (canonicalized internally).
	_, taken := IsStemTaken("balloon")
	uassert.True(t, taken, "raw stem should hit via internal canonicalization")

	// `baiioon` (the canonical form, with l→i applied) also finds it.
	_, taken2 := IsStemTaken("baiioon")
	uassert.True(t, taken2, "pre-canonicalized form should also hit")

	// `ballbon` (different stem, also has l) should NOT find it —
	// canonical is `baiibon` ≠ `baiioon`.
	_, taken3 := IsStemTaken("ballbon")
	uassert.False(t, taken3, "different stem should not collide")
}

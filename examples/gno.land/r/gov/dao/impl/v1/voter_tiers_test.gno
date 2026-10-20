package impl

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	implv0 "gno.land/r/gov/dao/impl/v0"
	"gno.land/r/gov/dao/memberstore/v0"
)

// impl/v1 enforces impl/v0's tier filters with v0's rules, and refuses any
// filter it cannot enforce instead of treating it as unrestricted.
func TestVoterTiers(t *testing.T) {
	all := "T1,T2,T3"
	cases := []struct {
		name   string
		filter any
		want   string // "" means refused
	}{
		{"no filter", nil, all},
		{"v1 T1", NewFilterByTier(memberstore.T1), "T1"},
		{"v1 T2", NewFilterByTier(memberstore.T2), "T1,T2"},
		{"v1 T3", NewFilterByTier(memberstore.T3), all},
		{"v0 T1", implv0.NewFilterByTier(memberstore.T1), "T1"},
		{"v0 T2", implv0.NewFilterByTier(memberstore.T2), "T1,T2"},
		{"v0 T3", implv0.NewFilterByTier(memberstore.T3), all},
		{"unknown tier", NewFilterByTier("T9"), ""},
		{"unknown v0 tier", implv0.NewFilterByTier(""), ""},
		{"foreign type", struct{ Tier string }{"T1"}, ""},
		{"pointer to v1 filter", &FilterByTier{Tier: "T1"}, ""},
	}
	for _, c := range cases {
		got, err := voterTiers(c.filter)
		if c.want == "" {
			uassert.ErrorIs(t, err, ErrUnsupportedFilter, c.name)
			continue
		}
		uassert.NoError(t, err, c.name)
		uassert.Equal(t, c.want, strings.Join(got, ","), c.name)
	}
}

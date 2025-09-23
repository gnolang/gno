package foo721

import (
	"testing"

	"gno.land/p/demo/tokens/grc721"
)

func TestFoo721(t *testing.T) {
	admin := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	hariom := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	for _, tc := range []struct {
		name     string
		expected any
		fn       func() any
	}{
		{"BalanceOf(admin)", int64(10), func() interface{} { return BalanceOf(admin) }},
		{"BalanceOf(hariom)", int64(5), func() interface{} { return BalanceOf(hariom) }},
		{"OwnerOf(0)", admin, func() interface{} { return OwnerOf(grc721.TokenID("0")) }},
		{"IsApprovedForAll(admin, hariom)", false, func() interface{} { return IsApprovedForAll(admin, hariom) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if tc.expected != got {
				t.Errorf("expected: %v got: %v", tc.expected, got)
			}
		})
	}
}

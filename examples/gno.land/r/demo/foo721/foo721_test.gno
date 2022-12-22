package foo721

import (
	"testing"

	"gno.land/p/demo/grc/grc721"
	"gno.land/r/demo/users"
)

func TestFoo721(t *testing.T) {
	admin := users.AddressOrName("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	hariom := users.AddressOrName("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	for i, tc := range []struct {
		name     string
		expected interface{}
		fn       func() interface{}
	}{
		{"BalanceOf(admin)", 10, func() interface{} { return BalanceOf(admin) }},
		{"BalanceOf(hariom)", 5, func() interface{} { return BalanceOf(hariom) }},
		{"OwnerOf(0)", admin.Resolve(), func() interface{} { return OwnerOf(grc721.TokenID("0")) }},
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

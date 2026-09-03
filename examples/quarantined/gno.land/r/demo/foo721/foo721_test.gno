package foo721

import (
	"testing"

	"gno.land/p/demo/tokens/grc721"
)

func TestFoo721(t *testing.T) {
	admin := address("g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf")

	for _, tc := range []struct {
		name     string
		expected any
		fn       func() any
	}{
		{"BalanceOf(admin)", int64(10), func() any { return BalanceOf(admin) }},
		{"OwnerOf(0)", admin, func() any { return OwnerOf(grc721.TokenID("0")) }},
		{"IsApprovedForAll(admin, admin)", false, func() any { return IsApprovedForAll(admin, admin) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if tc.expected != got {
				t.Errorf("expected: %v got: %v", tc.expected, got)
			}
		})
	}
}

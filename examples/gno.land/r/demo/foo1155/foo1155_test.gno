package foo1155

import (
	"testing"

	"gno.land/p/demo/tokens/grc1155"
)

func TestFoo721(t *testing.T) {
	admin := address("g10x5phu0k6p64cwrhfpsc8tk43st9kug6wft530")
	bob := address("g1ze6et22ces5atv79y4xh38s4kuraey4y2fr6tw")
	tid1 := grc1155.TokenID("1")

	for _, tc := range []struct {
		name     string
		expected any
		fn       func() any
	}{
		{"BalanceOf(admin, tid1)", int64(100), func() any { return BalanceOf(admin, tid1) }},
		{"BalanceOf(bob, tid1)", int64(0), func() any { return BalanceOf(bob, tid1) }},
		{"IsApprovedForAll(admin, bob)", false, func() any { return IsApprovedForAll(admin, bob) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if tc.expected != got {
				t.Errorf("expected: %v got: %v", tc.expected, got)
			}
		})
	}
}

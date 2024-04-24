package foo1155

import (
	"testing"

	"gno.land/p/demo/grc/grc1155"
	"gno.land/p/demo/users"
)

func TestFoo721(t *testing.T) {
	admin := users.AddressOrName("g10x5phu0k6p64cwrhfpsc8tk43st9kug6wft530")
	bob := users.AddressOrName("g1ze6et22ces5atv79y4xh38s4kuraey4y2fr6tw")
	tid1 := grc1155.TokenID("1")
	tid2 := grc1155.TokenID("2")

	for i, tc := range []struct {
		name     string
		expected interface{}
		fn       func() interface{}
	}{
		{"BalanceOf(admin, tid1)", uint64(100), func() interface{} { return BalanceOf(admin, tid1) }},
		{"BalanceOf(bob, tid1)", uint64(0), func() interface{} { return BalanceOf(bob, tid1) }},
		{"IsApprovedForAll(admin, bob)", false, func() interface{} { return IsApprovedForAll(admin, bob) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fn()
			if tc.expected != got {
				t.Errorf("expected: %v got: %v", tc.expected, got)
			}
		})
	}
}

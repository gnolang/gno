package coins

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

func TestBalanceChecker(t *testing.T) {
	denom := "testtoken"
	addr1 := testutils.TestAddress("user1")
	addr2 := testutils.TestAddress("user2")

	realm := std.NewCodeRealm("gno.land/r/gnoland/coins")
	testing.SetRealm(realm)

	testing.IssueCoins(addr1, std.Coins{{denom, 1000000}})
	testing.IssueCoins(addr2, std.Coins{{denom, 500000}})

	tests := []struct {
		name      string
		path      string
		contains  string
		wantPanic bool
	}{
		{
			name:     "homepage",
			path:     "",
			contains: "# Gno.land Coins Explorer",
		},
		// TODO: not supported yet
		// {
		// 	name:     "total supply",
		// 	path:     denom,
		// 	expected: "Balance: 1500000testtoken",
		// },
		{
			name:     "addr1's balance",
			path:     ufmt.Sprintf("%s/%s", addr1.String(), denom),
			contains: ufmt.Sprintf("`%s` has `%d%s`", addr1.String(), 1000000, denom),
		},
		{
			name:     "addr2's balance",
			path:     ufmt.Sprintf("%s/%s", addr2.String(), denom),
			contains: ufmt.Sprintf("`%s` has `%d%s`", addr2.String(), 500000, denom),
		},
		//{
		//	name:      "invalid path",
		//	path:      ufmt.Sprintf("%s/invalid/extra", denom),
		//	contains:  "404",
		//	wantPanic: false,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic for %s", tt.name)
					}
				}()
			}

			result := Render(tt.path)
			if !tt.wantPanic {
				if !strings.Contains(result, tt.contains) {
					t.Errorf("expected %s to contain %s", result, tt.contains)
				}
			}
		})
	}
}
